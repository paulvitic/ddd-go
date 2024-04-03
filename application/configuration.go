package application

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type Configuration struct {
	Profile string
	Port    string `json:"port"`
}

func NewConfig(profile string) (config *Configuration, errs []error) {

	if profile != "" {
		logConfigInfo("starting in %s profile", profile)
		profile = "." + strings.ToLower(profile)
	} else {
		logConfigInfo("starting in production mode")
	}

	configPath := "./properties" + profile + ".json"
	logConfigInfo("loading configuration from %s", configPath)

	if configFile, err := os.Open(configPath); err != nil {
		log.Printf("%#v", err.Error())
		return nil, append(errs, err)

	} else {
		defer closeWhenDone(configFile, errs)()

		jsonParser := json.NewDecoder(configFile)

		config := Configuration{
			Profile: profile,
		}

		if err = jsonParser.Decode(&config); err != nil {
			log.Printf("%#v", err.Error())
			return nil, append(errs, err)
		}

		return &config, errs
	}
}

func closeWhenDone(target io.Closer, errs []error) func() {
	return func() {
		err := target.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}
}

func logConfigInfo(msg string, args ...interface{}) {
	log.Printf(fmt.Sprintf(fmt.Sprintf("[info] AppConfig: %s", msg), args...))
}
