package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/madhurendhar/CLASSPRO/helpers/databases"
	"github.com/madhurendhar/CLASSPRO/globals"
	"github.com/madhurendhar/CLASSPRO/handlers"
	"github.com/madhurendhar/CLASSPRO/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)


func main() {
	if globals.DevMode {
		_ = godotenv.Load()
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	app := fiber.New(fiber.Config{
		Prefork:      false,
		ServerHeader: "CLASSPRO",
		AppName:      "ClassPro API",
		JSONEncoder:  json.Marshal,
		JSONDecoder:  json.Unmarshal,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return utils.HandleError(c, err)
		},
	})

	// Middleware
	app.Use(recover.New())
	app.Use(compress.New(compress.Config{Level: compress.LevelBestSpeed}))
	app.Use(etag.New())

	allowedOrigins := "http://localhost:" + port
	if urls := os.Getenv("URL"); urls != "" {
		allowedOrigins += "," + urls
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,X-CSRF-Token,Authorization",
		ExposeHeaders:    "Content-Length",
		AllowCredentials: true,
	}))

	app.Use(limiter.New(limiter.Config{
		Max:        25,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			token := c.Get("X-CSRF-Token")
			if token != "" {
				return utils.Encode(token)
			}
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded. Please try again later.",
			})
		},
		SkipFailedRequests: false,
		LimiterMiddleware:  limiter.SlidingWindow{},
	}))

	// Setup Routes
	handlers.SetupRoutes(app)

	// Connect to Database
	if err := databases.Connect(); err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// Start Server
	log.Printf("Server is running on port %s 🚀", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
