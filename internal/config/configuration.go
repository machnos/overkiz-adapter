package config

import (
	"encoding/json"
	gpv "github.com/go-playground/validator/v10"
	"os"
)

type Configuration struct {
	Region   string `json:"region" validate:"required,oneof='europe' 'middle east' 'africa' 'asia' 'pacific' 'north america'"`
	Pod      string `json:"pod" validate:"required"`
	UserID   string `json:"user_id" validate:"required"`
	Password string `json:"password" validate:"required"`
	Http     *Http  `json:"http" validate:"required"`
}

type Http struct {
	Port        uint16 `json:"port"`
	ContextRoot string `json:"context_root"`
}

func LoadConfiguration(configFile string) (*Configuration, error) {
	file, err := os.Open(configFile)
	if err != nil {
		defer func(file *os.File) {
			_ = file.Close()
		}(file)
		return nil, err
	}

	decoder := json.NewDecoder(file)
	configuration := &Configuration{}
	err = decoder.Decode(configuration)
	if err != nil {
		return nil, err
	}
	validator := gpv.New()
	err = validator.Struct(configuration)
	if err != nil {
		return nil, err
	}
	return configuration, nil
}
