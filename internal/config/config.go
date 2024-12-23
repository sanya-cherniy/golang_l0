package config

import (
	"l0/pkg/logging"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Listen struct {
		Type   string `yaml:"type" env-default:"port"`
		BindIP string `yaml:"bind_ip" env-default:"127.0.0.1"`
		Port   string `yaml:"port" env-default:"8080"`
	} `yaml:"listen"`
	Storage    `yaml:"storage"`
	Brokers    []string `yaml:"brokers" env-default:"localhost:9092"`
	Containers string   `yaml:"containers"`
	LifeTime   int64    `yaml:"life_time" env-default:"10000"`
}

type Storage struct {
	Host     string `yaml:"host" env-default:"postgres"`
	Port     string `yaml:"port" env-default:"5432"`
	Username string `yaml:"username" env-default:"postgres"`
	Password string `yaml:"password" env-default:"postgres"`
	DBName   string `yaml:"db_name" env-default:"postgres"`
}

type Containers struct {
	Server   string `yaml:"server" env-default:"server"`
	Consumer string `yaml:"consumer" env-default:"consumer"`
	Broker   string `yaml:"broker" env-default:"kafka"`
	DataBase string `yaml:"data_base" env-default:"postgres"`
}

var instance *Config
var once sync.Once

func GetConfig(logFile string) *Config {
	once.Do(func() {
		logger, err := logging.GetLogger(logFile)
		if err != nil {
			panic(err)
		}
		logger.Info("read application configuration")
		instance = &Config{}
		if err := cleanenv.ReadConfig("config.yaml", instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			logger.Info(help)
			logger.Fatal(err)
		}

	})
	return instance
}
