package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	// Initialize database
	db, err := InitDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Create Fiber app
	app := fiber.New()

	// Enable CORS - Allow all origins for development
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, PUT, POST, DELETE, OPTIONS, HEAD",
		AllowCredentials: false,
		ExposeHeaders:    "Content-Length",
		MaxAge:           3600,
	}))

	// API routes
	app.Get("/api/content/:id", GetContent(db))
	app.Put("/api/content/:id", PutContent(db))
	app.Options("/api/content/:id", func(c *fiber.Ctx) error {
		return c.SendStatus(204)
	})

	// Start server
	port := ":9000"
	log.Printf("Server started on %s\n", port)
	log.Fatal(app.Listen(port))
}
