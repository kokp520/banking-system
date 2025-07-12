package main

import (
	"github.com/gin-gonic/gin"
	"github.com/kokp520/banking-system/server/internal/middleware"
	"log"
	"net/http"

	"github.com/kokp520/banking-system/server/internal/handler"
	"github.com/kokp520/banking-system/server/internal/service"
	"github.com/kokp520/banking-system/server/pkg/config"
	"github.com/kokp520/banking-system/server/pkg/logger"
	"github.com/kokp520/banking-system/server/pkg/storage"
)

var cfg *config.Config

func init() {
	var err error
	if cfg, err = config.Setup(); err != nil {
		log.Fatal("failed to load config", err)
	}
	if err := logger.Init("info", "json", "logs"); err != nil {
		log.Fatal("failed to init logger", err)
	}
}

func main() {
	gin.SetMode(cfg.Server.Mode)
	r := initRouter()
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

// 可擴充性說明：
// middleware：jwt、rate limit、cors etc.
// 依賴注入：DI, todo: unit test and integration test
// restful api 原則
func initRouter() *gin.Engine {
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(middleware.Logger())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	memoryStorage := storage.NewMemoryStorage()
	accountService := service.NewAccountService(memoryStorage)
	accountHandler := handler.NewAccountHandler(accountService)

	v1 := r.Group("/v1")
	{
		accounts := v1.Group("/accounts")
		{
			accounts.POST("", accountHandler.CreateAccount)
		}
	}

	return r
}
