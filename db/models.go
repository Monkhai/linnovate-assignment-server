package db

import (
	"time"
)

// Product represents a product in the database
// Product represents a product for testing purposes
type Product struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Price     float64   `db:"price"`
	CreatedAt time.Time `db:"created_at"`
}

// Review represents a review for testing purposes
type Review struct {
	ID        int64     `db:"id"`
	Content   string    `db:"content"`
	Rating    float64   `db:"rating"`
	ProductID int64     `db:"product_id"`
	CreatedAt time.Time `db:"created_at"`
}

// ProductReview represents a product-review relationship for testing purposes
type ProductReview struct {
	ID        int64     `db:"id"`
	ProductID int64     `db:"product_id"`
	ReviewID  int64     `db:"review_id"`
	CreatedAt time.Time `db:"created_at"`
}
