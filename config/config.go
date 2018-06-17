package config

import (
	"github.com/BurntSushi/toml"
)

type Config struct {
	Nginx    Nginx
	Database Database
	Ingest   Ingest
	Digest   Digest
}

type Nginx struct {
	Format string
}

type Database struct {
	File string
}

type Ingest struct {
	PiggyBackDigest bool
}

type Digest struct {
	RbfsLayerCap        int
	MinPathPermutations int
	UriRegex            string
}

const configPath string = "./log-analyse.toml"

func GetConfig() (Config, error) {
	config := Config{}
	_, parseError := toml.DecodeFile(configPath, &config)

	return config, parseError
}
