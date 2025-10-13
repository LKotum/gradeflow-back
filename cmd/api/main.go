// @title       GradeFlow API
// @version     0.1.0
// @description API for schedule, assessments, attendance and grades.
// @BasePath    /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter the token with the `Bearer ` prefix, e.g. `Bearer <token>`
package main

import (
	"log"
	"os"

	"gradeflow/internal/app"
	docs "gradeflow/pkg/swagger"
)

func init() {
	// Можно настраивать метаданные тут, если нужно
	docs.SwaggerInfo.BasePath = "/api"
}

func main() {
	pgURL := mustEnv("PG_URL")
	httpAddr := getenv("HTTP_ADDR", ":8080")
	a := app.New(pgURL)
	srv := a.Engine
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
