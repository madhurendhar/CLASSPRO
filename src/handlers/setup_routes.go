package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// SetupRoutes initializes all the routes in the Fiber app
func SetupRoutes(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Welcome to GoScraper API!"})
	})

	// Add more routes as needed
}
