// cmd/api/main.go
package main

import (
	"context"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	edb "gradeflow/internal/db"
	httpapi "gradeflow/internal/http"
)

func main() {
	pgURL := mustEnv("PG_URL")
	httpAddr := getenv("HTTP_ADDR", ":8080")

	gdb, err := gorm.Open(postgres.Open(pgURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("gorm open: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// ✔ авто-инициализация БД (версионированные миграции up)
	if err := edb.AutoMigrateUp(ctx, gdb); err != nil {
		log.Fatalf("migrate up: %v", err)
	}

	// старт API
	srv := httpapi.NewServer()
	log.Printf("API listening on %s", httpAddr)
	if err := srv.Run(httpAddr); err != nil {
		log.Fatal(err)
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("missing %s", k)
	}
	return v
}
