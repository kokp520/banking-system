package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"

	"github.com/kokp520/banking-system/server/pkg/config"
)

var cfg *config.Config

func init() {
	var err error
	if cfg, err = config.Setup(); err != nil {
		log.Fatal("failed to load config", err)
	}
}

func main() {
	gin.SetMode(cfg.Server.Mode)
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
