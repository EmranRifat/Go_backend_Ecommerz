package main

import (
	"fmt"
	"go-fiber-api/config"
	"go-fiber-api/database"
	"go-fiber-api/logger"
	"go-fiber-api/middleware"
	"go-fiber-api/routes"
	"go-fiber-api/security"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
)

func main() {

	// Load environment variables properly
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️.env file not loaded")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Config load failed:", err)
	}
	jwtm := security.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiresHours)

	// DB connect
	db, err := database.ConnectDB()
	if err != nil {
		logger.Error("Failed to connect to database", err)
		return
	}

	if err := database.Ping(db); err != nil {
		logger.Error("DB ping failed", err)
		return
	}

	logger.Success("DB Connection OK 👍")

	// Setup HTML Engine
	engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{
		AppName: "Go Fiber API",
		Views:   engine,
	})

	app.Static("/uploads", "./uploads")
	app.Use(fiberlogger.New())
	app.Use(recover.New())

	// app.Use(cors.New(cors.Config{
	// 	AllowOrigins: os.Getenv("FRONTEND_URL"),
	// 	AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	// 	AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
	// 	AllowCredentials: true,
	// }))
	app.Use(cors.New(cors.Config{
    AllowOrigins: "https://homelystay-frontend.vercel.app,http://localhost:3000",
    AllowHeaders: "Origin, Content-Type, Accept, Authorization",
    AllowMethods: "GET,POST,PUT,DELETE,PATCH,OPTIONS",
    AllowCredentials: true,
}))
	app.Use("/api", middleware.ActivityLogger(db))

	if err := os.MkdirAll("uploads", 0755); err != nil {
		log.Fatal("Failed to create uploads directory:", err)
	}
	app.Static("/uploads", "./uploads")

	// Routes
	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// ✅ Home route to render HTML
	// app.Get("/", func(c *fiber.Ctx) error {
	// 	return c.Render("index", fiber.Map{})
	// })

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("🚀 Go Fiber API running...")
	})
	
	app.Get("/api/db/ping", func(c *fiber.Ctx) error {
		if err := database.Ping(db); err != nil {
			return c.Status(500).JSON(fiber.Map{"db": "down", "detail": err.Error()})
		}
		return c.JSON(fiber.Map{"db": "ok"})
	})

	routes.ManageRoutes(app, jwtm, db)
	routes.PaymentRoutes(app, db)

	addr := fmt.Sprintf("0.0.0.0:%s", cfg.AppPort)
	// app.Listen(addr)

	logger.Success(fmt.Sprintf("🚀 Server running on http://%s", addr))

	if err := app.Listen(addr); err != nil {
		logger.Error("Failed to start server", err)
	}
}
