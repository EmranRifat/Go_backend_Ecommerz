// controllers/booking.go
package controllers

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"go-fiber-api/models"
	"go-fiber-api/types"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// bookingUserIDFromContext resolves the authenticated user ID set by middleware.Protect.
func bookingUserIDFromContext(c *fiber.Ctx) (uint, error) {
	sub, ok := c.Locals("sub").(string)
	if !ok || strings.TrimSpace(sub) == "" {
		return 0, errors.New("missing user subject")
	}

	id, err := strconv.ParseUint(sub, 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New("invalid user subject")
	}

	return uint(id), nil
}

func CreateBookingDB(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var in types.CreateBookingRequest

		// Parse request body
		if err := c.BodyParser(&in); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid body",
			})
		}

		// Validation
		if strings.TrimSpace(in.ListingID) == "" {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"error": "listing_id is required",
			})
		}

		if in.PaymentMethod == "" {
			in.PaymentMethod = "manual"
		}

		if in.Adults < 1 {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"error": "at least 1 adult is required",
			})
		}

		if in.TotalAmount < 0 {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"error": "total_amount must be >= 0",
			})
		}

		if strings.TrimSpace(in.Currency) == "" {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"error": "currency is required",
			})
		}

		if !in.TermsAccepted {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"error": "terms must be accepted",
			})
		}

		// Parse dates
		checkIn, err := time.Parse("2006-01-02", in.CheckIn)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid check_in format, use YYYY-MM-DD",
			})
		}

		checkOut, err := time.Parse("2006-01-02", in.CheckOut)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid check_out format, use YYYY-MM-DD",
			})
		}

		if !checkOut.After(checkIn) {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"error": "check_out must be after check_in",
			})
		}

		// Get authenticated user ID
		userID, err := bookingUserIDFromContext(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		// Fetch authenticated user
		var user models.User
		if err := db.First(&user, userID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "user not found",
			})
		}

		// Check duplicate booking
		var existing models.Booking

		err = db.Where(
			"listing_id = ? AND booked_by_id = ? AND check_in = ? AND check_out = ?",
			in.ListingID,
			user.ID,
			checkIn,
			checkOut,
		).First(&existing).Error

		if err == nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "booking already exists",
			})
		}

		if err != nil && err != gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to check existing booking",
			})
		}

		// Create booking
		booking := models.Booking{
			BookingID:     string(in.BookingID),
			BookedByID:    user.ID,
			BookedByName:  user.Name,
			BookedByEmail: user.Email,
			BookedByRole:  user.Role,

			// Guest information (flattened from nested request object)
			UserName:  in.UserInformation.Name,
			UserEmail: in.UserInformation.Email,
			UserPhone: in.UserInformation.Phone,
			UserRole:  in.UserInformation.Role,

			ListingID:     in.ListingID,
			PaymentMethod: in.PaymentMethod,

			// Listing snapshot
			ProductTitle:   in.ProductTitle,
			ProductImage:   in.ProductImage,
			Category:       in.Category,
			ProductAddress: in.ProductAddress,

			CheckIn:  checkIn,
			CheckOut: checkOut,

			Adults:      in.Adults,
			Children:    in.Children,
			TotalAmount: in.TotalAmount,
			Currency:    in.Currency,

			CardLast4:      in.CardLast4,
			CardExpiration: in.CardExpiration,
			TermsAccepted:  in.TermsAccepted,
		
			Status: "confirmed",
		}

		// Save booking
		if err := db.Create(&booking).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to create booking",
			})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": "booking created successfully",
			"data":    booking,
		})
	}
}
