package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"

	"goscraper/src/globals"
	"goscraper/src/handlers"
	"goscraper/src/helpers/databases"
	"goscraper/src/types"
	"goscraper/src/utils"
)

func main() {
	if globals.DevMode {
		_ = godotenv.Load()
	}

	app := fiber.New(fiber.Config{
		Prefork:      os.Getenv("PREFORK") == "true",
		ServerHeader: "GoScraper",
		AppName:      "GoScraper v3.0",
		JSONEncoder:  json.Marshal,
		JSONDecoder:  json.Unmarshal,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return utils.HandleError(c, err)
		},
	})

	// Middleware
	app.Use(recover.New(), compress.New(), etag.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:243," + os.Getenv("URL"),
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-CSRF-Token",
		ExposeHeaders:    "Content-Length",
		AllowCredentials: true,
	}))

	app.Use(limiter.New(limiter.Config{
		Max:        25,
		Expiration: 1 * time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"error": "Rate limit exceeded. Try again later."})
		},
	}))

	// CSRF Protection Middleware
	app.Use(func(c *fiber.Ctx) error {
		if c.Path() == "/login" || c.Path() == "/hello" {
			return c.Next()
		}
		token := c.Get("X-CSRF-Token")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing X-CSRF-Token header"})
		}
		return c.Next()
	})

	// Authorization Middleware
	app.Use(func(c *fiber.Ctx) error {
		if globals.DevMode || c.Path() == "/hello" {
			return c.Next()
		}
		token := c.Get("Authorization")
		if token == "" || (!strings.HasPrefix(token, "Bearer ") && !strings.HasPrefix(token, "Token ")) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing or invalid Authorization header"})
		}
		return c.Next()
	})

	// Routes
	app.Get("/hello", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Hello, World!"})
	})

	app.Post("/login", func(c *fiber.Ctx) error {
		var creds struct {
			Username string `json:"account"`
			Password string `json:"password"`
		}
		if err := c.BodyParser(&creds); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON body"})
		}
		if creds.Username == "" || creds.Password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing account or password"})
		}
		lf := &handlers.LoginFetcher{}
		session, err := lf.Login(creds.Username, creds.Password)
		if err != nil {
			return err
		}
		return c.JSON(session)
	})

	// Fetch All Data Route
	app.Get("/get", cache.New(cache.Config{Expiration: 2 * time.Minute}), func(c *fiber.Ctx) error {
		token := c.Get("X-CSRF-Token")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing X-CSRF-Token header"})
		}
		data, err := fetchAllData(token)
		if err != nil {
			return utils.HandleError(c, err)
		}
		return c.JSON(data)
	})

	// Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on port %s...", port)
	if err := app.Listen("0.0.0.0:" + port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// Fetch All Data Function
func fetchAllData(token string) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	var results = []struct {
		key  string
		fetch func(string) (interface{}, error)
	}{
		{"user", handlers.GetUser},
		{"attendance", handlers.GetAttendance},
		{"marks", handlers.GetMarks},
		{"courses", handlers.GetCourses},
		{"timetable", handlers.GetTimetable},
	}

	for _, r := range results {
		if value, err := r.fetch(token); err == nil {
			data[r.key] = value
		} else {
			return nil, err
		}
	}
	return data, nil
}
