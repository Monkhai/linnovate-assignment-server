package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB represents the database connection pool
type DB struct {
	pool *pgxpool.Pool
}

type Product struct {
	ID          int64     `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Price       float64   `db:"price" json:"price"`
	Image       string    `db:"image" json:"image"`
	Description string    `db:"description" json:"description"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
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

type ClientReview struct {
	ProductID     int64   `json:"productId"`
	ReviewTitle   string  `json:"reviewTitle"`
	ReviewContent string  `json:"reviewContent"`
	Stars         float64 `json:"stars"`
}
type SafeReview struct {
	ID            int64   `json:"id"`
	ProductID     int64   `json:"productId"`
	ReviewTitle   string  `json:"reviewTitle"`
	ReviewContent string  `json:"reviewContent"`
	Stars         float64 `json:"stars"`
}

// New creates a new database connection pool
func New(databaseURL string) (*DB, error) {
	fmt.Println("Parsing database configuration...")
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		fmt.Printf("Failed to parse database URL: %v\n", err)
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}
	fmt.Println("Successfully parsed database configuration")

	fmt.Println("Creating connection pool...")
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		fmt.Printf("Failed to create connection pool: %v\n", err)
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}
	fmt.Println("Successfully created connection pool")

	fmt.Println("Testing database connection with ping...")
	if err := pool.Ping(context.Background()); err != nil {
		fmt.Printf("Failed to ping database: %v\n", err)
		fmt.Println("Closing connection pool due to failed ping")
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	fmt.Println("Successfully pinged database")

	fmt.Println("Database connection successfully established")
	return &DB{pool: pool}, nil
}

// Close closes the database connection pool
func (db *DB) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

// HealthCheck checks if the database connection is healthy
func (db *DB) HealthCheck(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

func (db *DB) GetProducts(ctx context.Context) ([]Product, error) {
	rows, err := db.pool.Query(ctx, "SELECT * FROM products ORDER BY id")
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

func (db *DB) GetProduct(ctx context.Context, id int64) (Product, error) {
	rows, err := db.pool.Query(ctx, "SELECT * FROM products WHERE id = $1", id)
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

func (db *DB) PostReview(ctx context.Context, review ClientReview, userId string) (Review, error) {
	var newReview Review
	err := db.pool.QueryRow(ctx,
		`INSERT INTO reviews (user_id, product_id, review_title, review_content, stars) 
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, product_id, review_title, review_content, stars, created_at`,
		userId, review.ProductID, review.ReviewTitle, review.ReviewContent, review.Stars).Scan(
		&newReview.ID, &newReview.UserId, &newReview.ProductID, &newReview.ReviewTitle,
		&newReview.ReviewContent, &newReview.Stars, &newReview.CreatedAt)
	if err != nil {
		return Review{}, fmt.Errorf("failed to insert review: %w", err)
	}
	return newReview, nil
}

func (db *DB) GetProductReviews(ctx context.Context, productId int64) ([]Review, error) {
	rows, err := db.pool.Query(ctx, `
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
