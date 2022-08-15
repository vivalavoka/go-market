package config

import (
	"flag"

	"github.com/caarlos0/env"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Address              string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
}

func Init() (Config, error) {
	var config Config
	flag.StringVar(&config.Address, "a", "127.0.0.1:8080", "server address")
	flag.StringVar(&config.DatabaseURI, "d", "", "database dsn")
	flag.StringVar(&config.AccrualSystemAddress, "r", "", "accrual system address")

	flag.Parse()

	err := env.Parse(&config)
	if err != nil {
		return config, err
	}

	log.Info(config)
	return config, nil
}
