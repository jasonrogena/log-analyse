package config

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Nginx    Nginx
	Database Database
}

type Nginx struct {
	Format string
}

type Database struct {
	File string
}

const configPath string = "./log-analyse.toml"

func GetConfig() (Config, error) {
	config := Config{}
	_, parseError := toml.DecodeFile(configPath, &config)

	return config, parseError
}
