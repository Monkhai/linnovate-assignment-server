package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	client *pgxpool.Pool
	ctx    context.Context
}

type Product struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Price     float64   `db:"price"`
	CreatedAt time.Time `db:"created_at"`
}

type Review struct {
	ID            int64     `db:"id"`
	UserId        string    `db:"user_id"`
	ProductID     int64     `db:"product_id"`
	ReviewTitle   string    `db:"review_title"`
	ReviewContent string    `db:"review_content"`
	Stars         float64   `db:"stars"`
	CreatedAt     time.Time `db:"created_at"`
}

type UserReview struct {
	UserId        string  `json:"userId"`
	ProductID     int64   `json:"productId"`
	ReviewTitle   string  `json:"reviewTitle"`
	ReviewContent string  `json:"reviewContent"`
	Stars         float64 `json:"stars"`
}

func New(ctx context.Context, client *pgxpool.Pool) *DB {
	return &DB{ctx: ctx, client: client}
}

func (db *DB) Close(ctx context.Context) {
	if db.client != nil {
		db.client.Close()
	}
}

func (db *DB) GetProducts() ([]Product, error) {
	rows, err := db.client.Query(db.ctx, "SELECT * FROM products ORDER BY $1", "id")
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		product, err := pgx.RowToStructByName[Product](rows)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, nil
}

func (db *DB) GetProduct(id int64) (Product, error) {
	rows, err := db.client.Query(db.ctx, "SELECT id, name, price, created_at FROM products WHERE id = $1", id)
	if err != nil {
		return Product{}, fmt.Errorf("failed to query product: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return Product{}, fmt.Errorf("product with id %d not found", id)
	}

	product, err := pgx.RowToStructByName[Product](rows)
	if err != nil {
		return Product{}, fmt.Errorf("failed to serialize product: %w", err)
	}
	return product, nil
}

func (db *DB) PostReview(review UserReview) error {
	_, err := db.client.Exec(db.ctx,
		`INSERT INTO reviews (user_id, product_id, review_title, review_content, stars) 
		VALUES ($1, $2, $3, $4, $5)`,
		review.UserId, review.ProductID, review.ReviewTitle, review.ReviewContent, review.Stars)
	if err != nil {
		return fmt.Errorf("failed to insert review: %w", err)
	}
	return nil
}

func (db *DB) GetProductReviews(productId int64) ([]Review, error) {
	rows, err := db.client.Query(db.ctx, `
	SELECT * FROM reviews WHERE product_id = $1
	`, productId)

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var reviews []Review
	for rows.Next() {
		review, err := pgx.RowToStructByName[Review](rows)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	return reviews, nil

}

// ===========================================
// =================HELPERS===================
// ===========================================

func (db *DB) getReview(id int64) (Review, error) {
	rows, err := db.client.Query(db.ctx, "SELECT * FROM reviews WHERE id = $1", id)
	if err != nil {
		return Review{}, fmt.Errorf("failed to query product: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return Review{}, fmt.Errorf("product with id %d not found", id)
	}

	review, err := pgx.RowToStructByName[Review](rows)
	if err != nil {
		return Review{}, fmt.Errorf("failed to serialize product: %w", err)
	}
	return review, nil
}
