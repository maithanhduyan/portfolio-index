package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	fiberws "github.com/gofiber/websocket/v2"
	"github.com/joho/godotenv"

	"portfolio-index/config"
	"portfolio-index/handlers"
	"portfolio-index/repository"
	"portfolio-index/services"
)

func main() {
	// Load .env in development
	_ = godotenv.Load()

	cfg := config.Load()

	// ─── Database ────────────────────────────────────────────────
	db, err := repository.NewTimescaleDB(cfg.DBUrl)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// ─── Redis ───────────────────────────────────────────────────
	cache := services.NewRedisCache(cfg.RedisUrl)
	defer cache.Close()

	// ─── Background Data Fetcher ─────────────────────────────────
	fetcher := services.NewFetcher(db, cache, cfg)
	go fetcher.Start()

	// ─── Fiber App ───────────────────────────────────────────────
	app := fiber.New(fiber.Config{
		AppName:               "Portfolio Index API v1.0",
		DisableStartupMessage: false,
		// Optimized for high concurrency
		Concurrency:     256 * 1024,
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
	})

	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} ${latency} ${method} ${path}\n",
	}))
	app.Use(compress.New(compress.Config{Level: compress.LevelBestSpeed}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.AllowedOrigins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET,POST",
	}))

	// ─── REST Routes ─────────────────────────────────────────────
	h := handlers.New(db, cache, cfg)
	api := app.Group("/api/v1")

	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "version": "1.0.0"})
	})
	api.Get("/indices", h.GetAllIndices)
	api.Get("/indices/:symbol", h.GetIndex)
	api.Get("/indices/:symbol/candles", h.GetCandles)
	api.Get("/indices/:symbol/stats", h.GetStats)
	api.Get("/indices/:symbol/ai-analysis", h.GetAIAnalysis)

	// ─── WebSocket ───────────────────────────────────────────────
	wsHub := handlers.NewHub(cache)
	go wsHub.Run()

	app.Use("/ws", func(c *fiber.Ctx) error {
		if fiberws.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws/prices", fiberws.New(wsHub.HandleConnection))

	// ─── Graceful Shutdown ───────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down server...")
		_ = app.Shutdown()
	}()

	log.Printf("Server starting on port %s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
