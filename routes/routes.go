// routes/routes.go
package routes

import (
	"go-fiber-api/controllers"
	"go-fiber-api/controllers/dashboard"
	"go-fiber-api/handlers"
	"go-fiber-api/middleware"
	"go-fiber-api/security"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func ManageRoutes(app *fiber.App, jwtm *security.JWTManager, db *gorm.DB) {

	api := app.Group("/api")

	// ==========================================================
	// Public Authentication
	// ==========================================================
	auth := api.Group("/auth")
	auth.Post("/register", controllers.RegisterDB(db))
	auth.Post("/login", controllers.LoginDB(jwtm, db))

	// ==========================================================
	// Public Listing APIs
	// ==========================================================
	api.Get("/listings", controllers.GetListingDataDB(db))
	api.Get("/listing/:id", controllers.GetListingByIDDB(db))
	api.Post("/product", controllers.CreateListingDB(db))

	// ==========================================================
	// Public Product Category APIs
	// ==========================================================
	api.Get("/product-categories", controllers.ProductCategories(db))
	api.Get("/product-categories/:id", controllers.SingleProductCategory(db))

	// ==========================================================
	// Booking APIs (Protected)
	// ==========================================================
	api.Post("/bookings", middleware.Protect(jwtm), controllers.CreateBookingDB(db))
	// ==========================================================
	// Orders
	// ==========================================================
	api.Get("/orders", handlers.GetAllOrders(db))
	api.Get("/orders/:id", handlers.GetOrderByID(db))

	// ==========================================================
	// Weather APIs
	// ==========================================================
	api.Get("/weather", controllers.ListWeatherDB(db))
	api.Get("/weather/:id", controllers.GetWeatherByIDDB(db))
	api.Get("/weather/division/:division", controllers.GetWeatherByDivisionDB(db))

	// ==========================================================
	// Protected User APIs
	// ==========================================================
	user := api.Group("/", middleware.Protect(jwtm))
	user.Get("/my-bookings", controllers.GetMyBookings(db))	


	user.Get("/me", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"subject": c.Locals("sub"),
			"email":   c.Locals("email"),
		})
	})
	
	user.Post("/host-listings", controllers.CreateHostListingHandler(db))

	// ==========================================================
	// Admin APIs
	// ==========================================================
	
	admin := api.Group("/admin",middleware.Protect(jwtm),middleware.RequireRoles("admin", "superadmin"))
	admin.Get("/allUsers", dashboard.GetAllUsers(db))
	admin.Get("/reviews", dashboard.GetAllReviews(db))
	admin.Get("/bookings", controllers.GetBookings(db))
	admin.Delete("/users/:id", dashboard.DeleteUser(db))
	admin.Patch("/users/:id/role", dashboard.UpdateUserRole(db))

	admin.Get("/host-listings", dashboard.GetAdminHostListingsHandler(db))

	admin.Patch(
	"/host-listings/:id/status",
	dashboard.UpdateHostListingStatusHandler(db),
)

	admin.Get("/all-logs", dashboard.GetAdminActivityLogsHandler(db))
}