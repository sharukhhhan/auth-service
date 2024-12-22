package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log"
	"path"
	"time"
)

type (
	Config struct {
		HTTP     HTTP     `yaml:"http"`
		Log      Log      `yaml:"log"`
		Database Database `yaml:"database"`
		JWT      JWT      `yaml:"jwt"`
		SMTP     SMTP     `yaml:"smtp"`
	}

	HTTP struct {
		Port            string        `env-required:"true" yaml:"port"`
		ShutdownTimeout time.Duration `env-default:"5s" yaml:"shutdown_timeout"`
	}

	Log struct {
		Level   string `yaml:"level"`
		LogPath string `env-default:"./logs" yaml:"log_path"`
	}

	JWT struct {
		SignKey         string        `env-required:"true" yaml:"sign_key"`
		TokenTTL        time.Duration `env-default:"20m" yaml:"token_ttl"`
		RefreshTokenTTL time.Duration `env-default:"168h" yaml:"refresh_token_ttl"`
	}

	Database struct {
		Postgres Postgres `yaml:"postgres"`
	}

	Postgres struct {
		Host          string `yaml:"host" env-required:"true"`
		Port          int    `yaml:"port" env-required:"true"`
		User          string `yaml:"user" env-required:"true"`
		Password      string `yaml:"password" env-required:"true"`
		Name          string `yaml:"name" env-required:"true"`
		MigrationPath string `yaml:"migration_path" env-default:"./migrations"`
	}

	SMTP struct {
		Host     string `yaml:"host" env-default:"smtp.example.com"`
		Port     int    `yaml:"port" env-default:"587"`
		User     string `yaml:"user" env-default:"sender@example.com"`
		Password string `yaml:"password" env-default:"password"`
	}
)

func NewConfig(configPath string) (*Config, error) {
	cfg := &Config{}

	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found")
	}

	err := cleanenv.ReadConfig(path.Join("./", configPath), cfg)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, fmt.Errorf("error reading env: %w", err)
	}

	return cfg, nil
}
