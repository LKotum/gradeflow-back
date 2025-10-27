package config

import (
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Cors        string        `env:"CORS" envDefault:"*"`
	AppURL      string        `env:"APP_URL,required"`
	JWTSecret   string        `env:"JWT_SECRET,required"`
	AccessTTL   time.Duration `env:"ACCESS_TTL" envDefault:"15m"`
	RefreshTTL  time.Duration `env:"REFRESH_TTL" envDefault:"720h"` // 30d
	PGURL       string        `env:"PG_URL,required"`
	HTTPAddr    string        `env:"HTTP_ADDR" envDefault:":8080"`
	APIBasePath string        `env:"API_BASE_PATH" envDefault:"/api"`
	MinIO       MinIOConfig
	Logging     LoggingConfig
}

type MinIOConfig struct {
	Endpoint      string `env:"MINIO_ENDPOINT" envDefault:"minio:9000"`
	AccessKey     string `env:"MINIO_ACCESS_KEY"`
	SecretKey     string `env:"MINIO_SECRET_KEY"`
	UseSSL        bool   `env:"MINIO_USE_SSL" envDefault:"false"`
	BucketAvatars string `env:"MINIO_BUCKET_AVATARS" envDefault:"avatars"`
	BucketReports string `env:"MINIO_BUCKET_REPORTS" envDefault:"reports"`
}

type LoggingConfig struct {
	Level      string `env:"LOG_LEVEL" envDefault:"info"`
	Path       string `env:"LOG_PATH"`
	MaxSizeMB  int    `env:"LOG_MAX_SIZE_MB" envDefault:"20"`
	MaxBackups int    `env:"LOG_MAX_BACKUPS" envDefault:"10"`
	MaxAgeDays int    `env:"LOG_MAX_AGE_DAYS" envDefault:"30"`
	AlsoStdout bool   `env:"LOG_STDOUT" envDefault:"true"`
}

func Load() Config {
	var c Config
	if err := env.Parse(&c); err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}
	return c
}
