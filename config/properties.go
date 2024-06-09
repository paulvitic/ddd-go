package config

import (
	"encoding/json"
	"errors"
	"github.com/getsops/sops/v3/decrypt"
	"log"
	"os"
)

func Properties[T any](profile ...string) *T {

	if fileName, fileNameErr := fileNameFor(profile...); fileNameErr != nil {
		panic(fileNameErr)
	} else {
		if filePath, isFileEncrypted, openFileErr := verifyFilePath(fileName); openFileErr != nil {
			panic(openFileErr)
		} else {
			if data, readDataErr := readData(filePath, isFileEncrypted); readDataErr != nil {
				panic(readDataErr)
			} else {
				config := new(T)
				err := json.Unmarshal(data, config)
				if err != nil {
					log.Printf("[error] AppServer: failed to parse config file %s: %s", filePath, err)
					panic(err)
				}

				return config
			}
		}
	}
}

func readData(filePath string, isFileEncrypted bool) ([]byte, error) {
	log.Printf("[info] AppServer: starting server with config file %s", filePath)
	if data, err := os.ReadFile(filePath); err != nil {
		log.Printf("[error] AppServer: failed to read config file %s: %s", filePath, err)
		return nil, err
	} else {
		if isFileEncrypted {
			if decryptedData, decryptErr := decryptData(data); decryptErr != nil {
				return nil, decryptErr
			} else {
				data = decryptedData
			}
		}
		return data, nil
	}
}

func decryptData(data []byte) ([]byte, error) {
	// Decrypt the data with SOPS
	decryptedData, decryptErr := decrypt.Data(data, "json")
	if decryptErr != nil {
		log.Printf("[error] AppServer: failed to decrypt config file: %s", decryptErr)
		return nil, decryptErr
	}
	return decryptedData, nil
}

func verifyFilePath(fileName string) (string, bool, error) {
	fileExt := ".json"
	encryptExt := ".enc"
	isFileEncrypted := false

	filePath := "./" + fileName + fileExt
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[warn] AppServer: failed to open config file %s: %s", filePath, err)
		filePath = "./" + fileName + encryptExt + fileExt
		file, err = os.Open(filePath)
		if err == nil {
			isFileEncrypted = true
		} else {
			log.Printf("[error] AppServer: failed to open config file %s: %s", filePath, err)
			return "", isFileEncrypted, err
		}
	}

	defer func(file *os.File) {
		closeErr := file.Close()
		if closeErr != nil {
			log.Printf("[error] AppServer: failed to close config file %s: %s", filePath, closeErr)
		}
	}(file)

	return filePath, isFileEncrypted, nil
}

func fileNameFor(profile ...string) (string, error) {
	name := "properties"

	if profile == nil {
		return name, nil
	}

	if len(profile) > 1 {
		return "", errors.New("only one profile suffix is allowed")
	}

	profileExt := profile[0]
	if profileExt == "" {
		return name, nil
	}

	return name + "." + profile[0], nil // Overwrite configPath with profile suffix
}
