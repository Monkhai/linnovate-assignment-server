package db

import (
	"time"
)

// NewProduct creates a new product with a default creation timestamp
func NewProduct(id int64, name string, price float64) Product {
	return Product{
		ID:        id,
		Name:      name,
		Price:     price,
		CreatedAt: time.Now().UTC(),
	}
}

// NewReview creates a new review with a default creation timestamp
func NewReview(id int64, content string, rating float64, productID int64) Review {
	return Review{
		ID:        id,
		Content:   content,
		Rating:    rating,
		ProductID: productID,
		CreatedAt: time.Now().UTC(),
	}
}

// NewProductReview creates a new product-review relationship with a default creation timestamp
func NewProductReview(id int64, productID int64, reviewID int64) ProductReview {
	return ProductReview{
		ID:        id,
		ProductID: productID,
		ReviewID:  reviewID,
		CreatedAt: time.Now().UTC(),
	}
}
