// @title       GradeFlow API
// @version     0.1.0
// @description GradeFlow management API.
// @BasePath    /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter the token with the `Bearer ` prefix, e.g. `Bearer <token>`
package main

import (
	"fmt"
	"os"

	"gradeflow/internal/app"
	"gradeflow/internal/config"
	"gradeflow/pkg/logger"
	docs "gradeflow/pkg/swagger"
)

func main() {
	cfg := config.Load()
	if _, err := logger.Init(cfg.Logging); err != nil {
		fmt.Fprintf(os.Stderr, "logger init: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	docs.SwaggerInfo.BasePath = cfg.APIBasePath

	application, err := app.New(cfg)
	if err != nil {
		logger.Fatal("app init failed", "error", err)
	}
	logger.Info("API listening", "addr", cfg.HTTPAddr)
	if err := application.Run(); err != nil {
		logger.Fatal("server stopped", "error", err)
	}
}
