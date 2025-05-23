package ddd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/getsops/sops/v3/decrypt"
)

func Configuration[T any](configPath ...string) (*T, error) {
	fileName := "configs/properties"
	if len(configPath) > 0 {
		if configPath[0] != "" {
			fileName = configPath[0]
		}
	}
	if filePath, isFileEncrypted, openFileErr := verifyFilePath(fileName); openFileErr != nil {
		return nil, openFileErr
	} else {
		if data, readDataErr := readData(filePath, isFileEncrypted); readDataErr != nil {
			panic(readDataErr)
		} else {
			config := new(T)
			if err := json.Unmarshal(data, config); err != nil {
				return nil, err
			}
			return config, nil
		}

	}
}

func verifyFilePath(fileName string) (string, bool, error) {
	fileExt := ".json"
	encryptExt := ".enc"
	isFileEncrypted := false

	for strings.HasPrefix(fileName, ".") {
		fileName = strings.TrimPrefix(fileName, ".")
	}

	for strings.HasPrefix(fileName, "/") {
		fileName = strings.TrimPrefix(fileName, "/")
	}

	fileName = strings.TrimSuffix(fileName, ".json")
	fileName = strings.TrimSuffix(fileName, ".enc")

	filePath := "./" + fileName + fileExt
	file, err := os.Open(filePath)
	if err != nil {
		filePath = "./" + fileName + encryptExt + fileExt
		file, err = os.Open(filePath)
		if err == nil {
			isFileEncrypted = true
		} else {
			return "", isFileEncrypted, err
		}
	}

	defer func(file *os.File) {
		closeErr := file.Close()
		if closeErr != nil {
			panic(fmt.Errorf("Failed to close config file %s: %s", filePath, closeErr))
		}
	}(file)

	return filePath, isFileEncrypted, nil
}

func readData(filePath string, isFileEncrypted bool) ([]byte, error) {
	if data, err := os.ReadFile(filePath); err != nil {
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
		return nil, decryptErr
	}
	return decryptedData, nil
}
