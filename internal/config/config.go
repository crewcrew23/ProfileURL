package config

import (
	"flag"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string `yaml:"env" env-required:"true"`
	Addr        string `yaml:"addr" env-required:"true"`
	StoragePath string `yaml:"storage_path" env-required:"true"`
	TokenTTL    string `yaml:"token_ttl" env-required:"true"`
	Secret      string `yaml:"secret" env-required:"true"`
}

func MustLoad() *Config {
	path := fetchConfiPath()
	return MustLoadByPath(path)
}

func MustLoadByPath(path string) *Config {
	if path == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exist:" + path)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("Fail to parse config")
	}

	return &cfg
}

func fetchConfiPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
