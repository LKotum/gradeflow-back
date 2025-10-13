package config

import (
	"log"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	AppURL     string        `env:"APP_URL,required"`
	JWTSecret  string        `env:"JWT_SECRET,required"`
	AccessTTL  time.Duration `env:"ACCESS_TTL" envDefault:"15m"`
	RefreshTTL time.Duration `env:"REFRESH_TTL" envDefault:"720h"` // 30d
	SMTP       SMTPConfig
	Redis      RedisConfig
	MinIO      MinIOConfig
}

type SMTPConfig struct {
	Host      string `env:"SMTP_HOST"`
	Port      int    `env:"SMTP_PORT" envDefault:"587"`
	User      string `env:"SMTP_USER"`
	Pass      string `env:"SMTP_PASS"`
	FromEmail string `env:"SMTP_FROM_EMAIL"`
	FromName  string `env:"SMTP_FROM_NAME" envDefault:"GradeFlow"`
}

type RedisConfig struct {
	Addr     string `env:"REDIS_ADDR" envDefault:"redis:6379"`
	Password string `env:"REDIS_PASSWORD"`
	DB       int    `env:"REDIS_DB" envDefault:"0"`
}

type MinIOConfig struct {
	Endpoint      string `env:"MINIO_ENDPOINT" envDefault:"minio:9000"`
	AccessKey     string `env:"MINIO_ACCESS_KEY"`
	SecretKey     string `env:"MINIO_SECRET_KEY"`
	UseSSL        bool   `env:"MINIO_USE_SSL" envDefault:"false"`
	BucketAvatars string `env:"MINIO_BUCKET_AVATARS" envDefault:"avatars"`
	BucketReports string `env:"MINIO_BUCKET_REPORTS" envDefault:"reports"`
}

func Load() Config {
	var c Config
	if err := env.Parse(&c); err != nil {
		log.Fatalf("load config: %v", err)
	}
	return c
}
