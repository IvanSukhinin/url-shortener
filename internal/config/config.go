package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string `yaml:"env" env-default:"dev"`
	AliasLength int    `yaml:"alias_length" env-default:"7"`
	HTTPServer  `yaml:"http_server"`
	Db          `yaml:"db"`
	SsoGrpcApi  `yaml:"sso_grpc_api"`
	App         `yaml:"app"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type Db struct {
	Host     string `yaml:"host" env-required:"true" env:"POSTGRES_HOST"`
	Port     string `yaml:"port" env-required:"true" env:"POSTGRES_PORT"`
	User     string `yaml:"user" env-required:"true" env:"POSTGRES_USER"`
	Password string `yaml:"password" env-required:"true" env:"POSTGRES_PASSWORD"`
	Db       string `yaml:"db" env-required:"true" env:"POSTGRES_DB"`
}

type SsoGrpcApi struct {
	Address      string        `yaml:"address" env-default:"localhost:44044"`
	Timeout      time.Duration `yaml:"timeout" env-default:"4s"`
	RetriesCount int           `yaml:"retries_count" env-default:"5"`
}

type App struct {
	Secret []byte `yaml:"secret" env-required:"true" env:"JWT_SECRET"`
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
