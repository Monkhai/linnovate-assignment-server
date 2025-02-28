package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	client *pgxpool.Pool
	ctx    context.Context
}

// New creates a DB connection using the provided client
func New(ctx context.Context, client *pgxpool.Pool) *DB {
	return &DB{ctx: ctx, client: client}
}

// Close closes the DB connection
func (db *DB) Close(ctx context.Context) {
	if db.client != nil {
		db.client.Close()
	}
}

func (db *DB) GetProducts() ([]Product, error) {
	rows, err := db.client.Query(db.ctx, "SELECT id, name, price, created_at FROM products ORDER BY $1", "id")
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
