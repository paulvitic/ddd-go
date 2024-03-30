package application

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"
)

type Configuration struct {
	Context  string
	Profile  string
	Port     string `json:"port"`
	RabbitMQ struct {
		Url string `json:"url"`
	} `json:"rabbitMq"`
	MongoDb struct {
		Uri string `json:"uri"`
	} `json:"mongoDb"`
}

func NewConfig(profile string) (config *Configuration, errs []error) {

	if profile != "" {
		log.Printf("[info] starting in %s profile", profile)
		profile = "." + strings.ToLower(profile)
	} else {
		log.Printf("[info] starting in production mode")
	}

	configPath := "./properties" + profile + ".json"
	log.Printf("[info] loading configuration from %s", configPath)

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
