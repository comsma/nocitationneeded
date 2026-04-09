package main

import (
	"log"
	"os"

	"github.com/comsma/nocitationneeded/internal/cache"
	"github.com/comsma/nocitationneeded/internal/config"
	"github.com/comsma/nocitationneeded/internal/scraper"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisClient, err := cache.NewRedisClient(redisAddr)
	if err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	citationCache := cache.NewCitationCache(redisClient)
	s := scraper.New()

	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Gzip())

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	h := NewHandler(e, citationCache, s, cfg)

	e.Static("/static", "ui/static")

	e.GET("/", h.GetHome)
	e.POST("/cite", h.PostCite)
	e.GET("/health", h.GetHealth)

	if err := e.Start(":8080"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
