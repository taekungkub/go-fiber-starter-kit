package main

import (
	"fmt"
	"go-fiber-stater-kit/config"
	"go-fiber-stater-kit/internal/api/ask"
	"go-fiber-stater-kit/internal/api/auth"
	"go-fiber-stater-kit/internal/api/user"
	"go-fiber-stater-kit/internal/database"
	"go-fiber-stater-kit/pkg/core"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/sashabaranov/go-openai"
)

func main() {
	cfg := config.LoadConfig()

	fmt.Println("APP_ENV:", cfg.APP_ENV)

	// Database connections
	db := database.ConnectPostgres(cfg)
	// redisStore := database.NewRedis(cfg)
	// redisClient := database.NewRedisClient(cfg)

	// JWT configuration
	core.AccessTokenSecret = []byte(cfg.JWTSecret)
	core.RefreshTokenSecret = []byte(cfg.RefreshSecret)
	core.AccessTokenExpiry = time.Duration(cfg.JWTExpiresInAccessToken) * time.Minute
	core.RefreshTokenExpiry = time.Duration(cfg.JWTExpiresInRefreshToken) * time.Minute

	client := openai.NewClient(cfg.OpenAIKey)

	// Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return core.SendError(c, code, err.Error())
		},
	})

	// Global middleware
	app.Use(cors.New())
	app.Use(logger.New())
	app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		// Storage:    redisStore,
	}))

	// Welcom
	app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.JSON(fiber.Map{
			"message": "Welcome to Go Fiber Starter Kit",
		})
	})

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})

	// API routes
	api := app.Group("/api")

	// Auth module
	authRepo := auth.NewRepository(db)
	authUseCase := auth.NewUseCase(authRepo)
	authHandler := auth.NewHandler(authUseCase)
	auth.AuthRouter(api, authHandler)

	// User module
	userRepo := user.NewRepository(db)
	userUseCase := user.NewUseCase(userRepo)
	userHandler := user.NewHandler(userUseCase)
	user.UserRouter(api, userHandler)

	// Bot module
	askRepo := ask.NewRepository(client)
	askUseCase := ask.NewUseCase(askRepo)
	askHandler := ask.NewHandler(askUseCase)
	ask.AskRouter(api, askHandler)

	log.Fatal(app.Listen(":8080"))
}
