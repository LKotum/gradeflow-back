package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewServer() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	return r
}
