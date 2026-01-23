package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string   `yaml:"env" env-default:"local"`
	RabbitMQ RabbitMQ `yaml:"rabbitmq"`
	SMTP     SMTP     `yaml:"-"`
	Email    Email    `yaml:"email"`
}

type RabbitMQ struct {
	URL     string `yaml:"url" env-required:"true"`
	Queue   string `yaml:"queue" env-required:"true"`
	Workers int    `yaml:"workers" env-default:"5"`
}

type SMTP struct {
	Host     string `env:"SMTP_HOST" env-required:"true"`
	Port     int    `env:"SMTP_PORT" env-required:"true"`
	Username string `env:"SMTP_USERNAME" env-required:"true"`
	Password string `env:"SMTP_PASSWORD" env-required:"true"`
}

type Email struct {
	From string `yaml:"from" env-required:"true"`
	Base string `yaml:"base" env-required:"true"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	if err := godotenv.Load(); err != nil {
		log.Fatalf("cannot load .env files: %s", err)
	}
	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("cannot read env: %s", err)
	}
	return &cfg
}
