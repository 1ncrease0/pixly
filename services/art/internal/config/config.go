package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env      string   `yaml:"env" env-default:"local"`
	Postgres Postgres `yaml:"postgres"`
	MinIO    MinIO    `yaml:"minio"`
	GRPC     GRPC     `yaml:"grpc"`
}

type MinIO struct {
	SecretKey  string        `yaml:"secret_key" env-required:"true"`
	AccessKey  string        `yaml:"access_key" env-required:"true"`
	UseSSL     bool          `yaml:"use_ssl"`
	Endpoint   string        `yaml:"endpoint" env-required:"true"`
	Buckets    []string      `yaml:"buckets" env-required:"true"`
	PresignTTL time.Duration `yaml:"presign_ttl" env-required:"true"`
}

type Postgres struct {
	DSN string `yaml:"dsn" env-required:"true"`
}

type GRPC struct {
	Port int `yaml:"port" env-default:"50052"`
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
