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
	ID        int64     `db:"id"`
	Content   string    `db:"content"`
	Rating    float64   `db:"rating"`
	ProductID int64     `db:"product_id"`
	CreatedAt time.Time `db:"created_at"`
}

type ProductReview struct {
	ID        int64     `db:"id"`
	ProductID int64     `db:"product_id"`
	ReviewID  int64     `db:"review_id"`
	CreatedAt time.Time `db:"created_at"`
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

	// Need to call Next() to position cursor at first row before reading
	if !rows.Next() {
		return Product{}, fmt.Errorf("product with id %d not found", id)
	}

	product, err := pgx.RowToStructByName[Product](rows)
	if err != nil {
		return Product{}, fmt.Errorf("failed to serialize product: %w", err)
	}
	return product, nil
}
