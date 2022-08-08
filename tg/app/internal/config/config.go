package config

import (
	"app/pkg/logging"
	"os"
	"path/filepath"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Level  string `yaml : "level"`
	Listen struct {
		Port string `yaml : "port"`
	} `yaml : "listen"`

	Tg struct {
		Token string `yaml : "token"`
	} `yaml : "tg"`

	Ipstack struct {
		Key string `yaml : "key"`
	} `yaml : "ipstack"`

	Rabbit struct {
		Host     string `yaml : "host"`
		Port     string `yaml : "port"`
		Username string `yaml : "username"`
		Password string `yaml : "password"`
		Consumer struct {
			Ipstack string `yaml : "name"`
			Buffer  int    `yaml : "buffer"`
		} `yaml : "consumer"`
		Producer struct {
			Name string `yaml : "name"`
		} `yaml : "producer"`
	} `yaml : "rabbit"`

	Event struct {
		WorkerCount int `yaml : "worker"`
	} `yaml : "event"`
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		wd, err := os.Getwd()
		if err != nil {
			// handle error
		}
		parentTop := filepath.Dir(wd)
		app := filepath.Dir(parentTop)
		instance = &Config{}
		logger := logging.GetLogger("info")
		logger.Info("start config initialisation")
		if err := cleanenv.ReadConfig(app+"\\config.yaml", instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			logger.Info(help)
			logger.Fatal(err)
		}

	})
	return instance
}
