package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func getConfigFile() (string, error) {
	homeDir, errHome := os.UserHomeDir()
	if errHome != nil {
		return "", errHome
	}

	filePath := filepath.Join(homeDir, configFileName)

	return filePath, nil
}

func getConfigStruct() (*Config, error) {
	filePath, errFile := getConfigFile()
	if errFile != nil {
		return nil, errFile
	}

	fileByte, errRead := os.ReadFile(filePath)
	if errRead != nil {
		return nil, errRead
	}

	config := &Config{}
	errUnmarshal := json.Unmarshal(fileByte, config)
	if errUnmarshal != nil {
		return nil, errUnmarshal
	}

	return config, nil
}

func writeConfig(cfg *Config) error {
	jsonBinary, errMarshal := json.Marshal(cfg)
	if errMarshal != nil {
		return errMarshal
	}

	cfgFile, errFile := getConfigFile()
	if errFile != nil {
		return errFile
	}

	errWrite := os.WriteFile(cfgFile, jsonBinary, 0644)
	
	return errWrite
}