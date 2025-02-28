package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DefaultMigrations contains the SQL to create the default schema for testing
var DefaultMigrations = []string{
	// 001 - Create products table
	`CREATE TABLE products (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		price DECIMAL(10, 2) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`,

	// 002 - Create reviews table
	`CREATE TABLE reviews (
		id SERIAL PRIMARY KEY,
		content TEXT NOT NULL,
		rating INTEGER NOT NULL CHECK (rating BETWEEN 1 AND 5),
		product_id INTEGER NOT NULL REFERENCES products(id),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`,

	// 003 - Create product_reviews table as a relationship table (not a view)
	`CREATE TABLE product_reviews (
		id SERIAL PRIMARY KEY,
		product_id INTEGER NOT NULL REFERENCES products(id),
		review_id INTEGER NOT NULL REFERENCES reviews(id),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(product_id, review_id)
	);`,
}

// ApplyDefaultMigrations applies the default set of migrations for testing
func applyDefaultMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	// Execute each migration as a separate transaction
	for i, migration := range DefaultMigrations {
		// Execute migration
		_, err := pool.Exec(ctx, migration)
		if err != nil {
			return fmt.Errorf("failed to execute migration #%d: %w", i+1, err)
		}
	}

	return nil
}
