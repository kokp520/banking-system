package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kokp520/banking-system/server/internal/middleware"
	"github.com/kokp520/banking-system/server/internal/storage"

	"github.com/kokp520/banking-system/server/internal/handler"
	"github.com/kokp520/banking-system/server/internal/service"
	"github.com/kokp520/banking-system/server/pkg/config"
	"github.com/kokp520/banking-system/server/pkg/logger"

	swaggerFiles "github.com/swaggo/files"     // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
)

var cfg *config.Config

func init() {

	var configFile string
	flag.StringVar(&configFile, "c", "", "config file path")
	flag.Parse()

	if configFile == "" {
		configFile = "config"
	}

	var err error
	if cfg, err = config.Setup(configFile); err != nil {
		log.Fatal("failed to load config", err)
	}
	if err := logger.Init(cfg.Logger.Level, cfg.Logger.Format, cfg.Logger.Dir); err != nil {
		log.Fatal("failed to init logger", err)
	}
}

func main() {
	gin.SetMode(cfg.Server.Mode)
	r := initRouter()

	// 未來可以改用httpserver, handler帶入gin engine, 支援更多可控性
	port := fmt.Sprintf(":%v", cfg.Server.Port)
	if e := r.Run(port); e != nil {
		log.Fatal("failed to start server", e)
	}
}

// 可擴充性說明：
// middleware：jwt、rate limit、cors etc.
// 依賴注入：DI, todo: unit test and integration test
// restful api 原則
func initRouter() *gin.Engine {
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.TraceID())

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
		account := v1.Group("/account")
		{
			account.POST("", accountHandler.CreateAccount)
			account.GET("/:id", accountHandler.GetAccount)
			account.POST("/:id/deposit", accountHandler.Deposit)
			account.POST("/:id/withdraw", accountHandler.Withdraw)
			account.POST("/:id/transfer", accountHandler.Transfer)
			account.GET("/:id/transactions", accountHandler.GetTransactions)
		}
	}

	// Swagger UI
	r.Static("/api", "./api")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL(cfg.Swagger.Path)))

	return r
}
