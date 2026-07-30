package models

import "time"

type Booking struct {
	ID uint `json:"id" gorm:"primaryKey"`

	// Client-provided booking id (optional, for traceability)
	BookingID string `json:"booking_id"`

	// Authenticated user snapshot
	BookedByID    uint   `json:"booked_by_id"`
	BookedByName  string `json:"booked_by_name"`
	BookedByEmail string `json:"booked_by_email"`
	BookedByRole  string `json:"booked_by_role"`

	// Guest information (flattened from request.user_information)
	UserName  string `json:"user_name"`
	UserEmail string `json:"user_email"`
	UserPhone string `json:"user_phone"`
	UserRole  string `json:"user_role"`

	// Listing
	ListingID     string `json:"listing_id"`
	PaymentMethod string `json:"payment_method"`

	// Listing snapshot (flattened from request.product_title/product_image/category/product_address)
	ProductTitle   string `json:"product_title"`
	ProductImage   string `json:"product_image"`
	Category       string `json:"category"`
	ProductAddress string `json:"product_address"`

	// Stay Details
	CheckIn  time.Time `json:"check_in"`
	CheckOut time.Time `json:"check_out"`

	Adults      int     `json:"adults"`
	Children    int     `json:"children"`
	TotalAmount float64 `json:"total_amount"`
	Currency    string  `json:"currency"`

	// Payment
	CardLast4      string `json:"card_last4"`
	CardExpiration string `json:"card_expiration"`
	TermsAccepted  bool   `json:"terms_accepted"`
	Status string `json:"status" gorm:"default:confirmed"`
	
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
