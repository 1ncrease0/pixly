package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Env          string       `yaml:"env" env-default:"local"`
	Postgres     Postgres     `yaml:"postgres"`
	GRPC         GRPC         `yaml:"grpc"`
	JWT          JWT          `yaml:"jwt"`
	Redis        Redis        `yaml:"redis"`
	Verification Verification `yaml:"verification"`
	RabbitMQ     RabbitMQ     `yaml:"rabbitmq"`
}

type JWT struct {
	Secret     string        `yaml:"secret_key" env-required:"true"`
	AccessTTL  time.Duration `yaml:"access_ttl" env-default:"5m"`
	RefreshTTL time.Duration `yaml:"refresh_ttl" env-default:"1h"`
}
type Postgres struct {
	DSN string `yaml:"dsn" env-required:"true"`
}

type GRPC struct {
	Port int `yaml:"port" env-default:"50051"`
}

type Redis struct {
	Addr     string `yaml:"addr" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
}

type Verification struct {
	TTL time.Duration `yaml:"ttl" env-default:"10m"`
}

type RabbitMQ struct {
	URL   string `yaml:"url" env-required:"true"`
	Queue string `yaml:"queue" env-required:"true"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
