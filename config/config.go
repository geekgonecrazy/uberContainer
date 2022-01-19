package config

import (
	"errors"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

var _config *config

type config struct {
	S3 S3Config `yaml:"s3"`
}

type S3Config struct {
	Endpoint         string `yaml:"endpoint"`
	Bucket           string `yaml:"bucket"`
	AccessKey        string `yaml:"accessKey"`
	AccessSecret     string `yaml:"accessSecret"`
	Region           string `yaml:"region"`
	UseSSL           bool   `yaml:"useSSL"`
	TempFileLocation string `yaml:"tempFileLocation"`
}

// Load tries to load the config file
func Load(filePath string) error {
	_config = new(config)

	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("yaml config read err  #%v ", err)
		return err
	}

	if err = yaml.Unmarshal(yamlFile, _config); err != nil {
		log.Fatalf("Unmarshal: %v", err)
		return err
	}

	return nil
}

// Get returns the config file
func Get() (*config, error) {
	if _config == nil {
		return nil, errors.New("no config loaded")
	}

	return _config, nil
}
