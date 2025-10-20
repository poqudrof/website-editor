package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	// Check log level
	if os.Getenv("LOG_LEVEL") == "HIGH" {
		log.Printf("🔍 [HIGH LOG] ================================")
		log.Printf("🔍 [HIGH LOG] HIGH LOGGING ENABLED")
		log.Printf("🔍 [HIGH LOG] All Claude API calls and responses will be logged in detail")
		log.Printf("🔍 [HIGH LOG] ================================")
	}

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

	// Content API routes
	app.Get("/api/content/:id", GetContent(db))
	app.Put("/api/content/:id", PutContent(db))
	app.Options("/api/content/:id", func(c *fiber.Ctx) error {
		return c.SendStatus(204)
	})

	// AI Command API routes (WebSocket-based)
	app.Post("/api/ai/command", ExecuteAICommand(db))
	app.Get("/api/ai/command/:commandId/stream", StreamAICommand(db))
	app.Get("/api/ai/command/:commandId/status", GetAICommandStatus(db))
	app.Post("/api/ai/command/:commandId/interrupt", InterruptAICommand())

	// Generic AI Agent API routes (SSE-based for custom CLI commands)
	app.Post("/api/agent/run", RunAgent())
	app.Get("/api/agent/stream/:sessionId", StreamAgent())
	app.Post("/api/agent/interrupt/:sessionId", InterruptAgent())
	app.Get("/api/agent/status/:sessionId", GetAgentStatus())
	app.Post("/api/agent/cleanup", CleanupSessions())

	// Start server
	port := ":9000"
	log.Printf("Server started on %s\n", port)
	log.Fatal(app.Listen(port))
}
