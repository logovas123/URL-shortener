package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// всё тоже самое как в yaml

type Config struct {
	Env         string `yaml:"env" env-default:"local"` // теги для считывания c yaml
	StoragePath string `yaml:"storage_path" env-requaired:"true"`
	HTTPServer  `yaml:"http_server"`
}

// добавили user и  password
type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
	User        string        `yaml:"user" env-required:"true"`
	Password    string        `yaml:"password" env-required:"true" env:"HTTP_SERVER_PASSWORD"`
}

// функция читает файл с конфигом и заполнит объект Config
// приставка Must по соглашению означает что функция не будет возвращать ошибку, а будет паниковать
func MustLoad() *Config {
	configPath := "/home/valera/Documents/Golang/MyProject/URL-shortener/config/local.yaml"
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	// проверяем существует ли файл
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	// читаем файл yaml, и извлекаем структуру Config в cfg
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
