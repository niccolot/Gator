package config

import "github.com/google/uuid"

const configFileName = ".gatorconfig.json"

type Config struct {
	DbURL string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
	CurrentUserID uuid.UUID `json:"current_user_id"`
	SuperUserName string `json:"superuser_name"`
	SuperUserID uuid.UUID `json:"superuser_id"`
	CmdHistory  []string `json:"cmd_history"`
}

func Read() *Config {
	config, errConfig := getConfigStruct()
	if errConfig != nil {
		return nil
	}

	return config
}

func (cfg *Config) SetUser(userName string, userID uuid.UUID) error {
	cfg.CurrentUserName = userName
	cfg.CurrentUserID = userID
	errWrite := writeConfig(cfg)
	
	return errWrite
}

