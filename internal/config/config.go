package config

const configFileName = ".gatorconfig.json"

type Config struct {
	DbURL string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() Config {
	config, errConfig := getConfigStruct()
	if errConfig != nil {
		return Config{}
	}

	return *config
}

func (cfg *Config) SetUser(userName string) error {
	cfg.CurrentUserName = userName
	errWrite := writeConfig(cfg)
	
	return errWrite
}

