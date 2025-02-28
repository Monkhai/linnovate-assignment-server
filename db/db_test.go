package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Fixed time for testing to ensure deterministic results
var testTime = time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

func setupEmptyDB(t *testing.T) (*DB, func()) {
	t.Helper()
	ctx := context.Background()

	db, err := NewTest(ctx)
	require.NoError(t, err)

	cleanup := func() {
		db.Close(ctx)
	}

	applyDefaultMigrations(ctx, db.client)
	return db, cleanup
}

func TestDB_Connection(t *testing.T) {
	db, cleanup := setupEmptyDB(t)
	defer cleanup()
	err := db.client.Ping(context.Background())
	assert.NoError(t, err)
}

func TestGetProducts(t *testing.T) {
	db, cleanup := setupEmptyDB(t)
	defer cleanup()

	ctx := context.Background()

	testProducts := []Product{
		{ID: 1, Name: "Test Product 1", Price: 19.99},
		{ID: 2, Name: "Test Product 2", Price: 29.99},
		{ID: 3, Name: "Test Product 3", Price: 39.99},
	}

	err := db.populateTestData(ctx, "products", testProducts)
	require.NoError(t, err)

	products, err := db.GetProducts()
	require.NoError(t, err)

	require.Len(t, products, len(testProducts))
	for i, p := range products {
		tp := testProducts[i]
		assert.Equal(t, tp.ID, p.ID)
		assert.Equal(t, tp.Name, p.Name)
		assert.Equal(t, tp.Price, p.Price)
	}
}
