package main

import (
	"log"
	"mcmanager/internal/config"
	"mcmanager/internal/http/manager"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	mgr := manager.New(cfg)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.GET("/healthz", func(c *gin.Context) {
		c.String(200, "ok")
	})
	router.POST("/update", mgr.UpdateHandler)

	log.Printf("listening on %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("http server failed: %v", err)
	}
}
