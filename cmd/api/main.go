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
	"fmt"
	"os"

	"gradeflow/internal/app"
	"gradeflow/internal/config"
	"gradeflow/pkg/logger"
	docs "gradeflow/pkg/swagger"
)

func main() {
	// Load config once to configure Swagger BasePath
	cfg := config.Load()
	if _, err := logger.Init(cfg.Logging); err != nil {
		fmt.Fprintf(os.Stderr, "logger init: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()
	docs.SwaggerInfo.BasePath = cfg.APIBasePath

	// Build app with already loaded configuration
	a := app.New(cfg)
	srv := a.Engine
	addr := a.Cfg.HTTPAddr
	logger.Info("API listening", "addr", addr)
	if err := srv.Run(addr); err != nil {
		logger.Fatal("server stopped", "error", err)
	}
}
