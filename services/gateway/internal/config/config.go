package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Env     string  `yaml:"env" env-default:"local"`
	HTTP    HTTP    `yaml:"http_server"`
	Clients Clients `yaml:"clients"`
}

type Clients struct {
	Auth GRPCService `yaml:"auth"`
}

type GRPCService struct {
	Addr    string        `yaml:"addr" env-required:"true"`
	Timeout time.Duration `yaml:"timeout" env-default:"5s"`
	Retries int           `yaml:"retries" env-default:"1"`
}

type HTTP struct {
	Host            string        `yaml:"host" env-default:"localhost"`
	Port            int           `yaml:"port"  env-default:"4000"`
	ReadTimeout     time.Duration `yaml:"read_timeout"  env-default:"5s"`
	WriteTimeout    time.Duration `yaml:"write_timeout"  env-default:"5s"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" env-default:"120s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"  env-default:"5s"`
	CORS            CORS          `yaml:"cors"`
}

type CORS struct {
	Enabled          bool          `yaml:"enabled" env-default:"false"`
	AllowedOrigins   []string      `yaml:"allowed_origins"`
	AllowedMethods   []string      `yaml:"allowed_methods"`
	AllowedHeaders   []string      `yaml:"allowed_headers"`
	AllowCredentials bool          `yaml:"allow_credentials"`
	MaxAge           time.Duration `yaml:"max_age"`
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
