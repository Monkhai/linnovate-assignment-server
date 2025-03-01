package server

import (
	"catalogapi/db"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*db.DB, func()) {
	t.Helper()
	ctx := context.Background()

	os.Setenv("APP_ENV", "test")
	database, err := db.NewTestDB(ctx)
	require.NoError(t, err)

	cleanup := func() {
		db.CloseTestDB(ctx, database)
	}
	// Set up test schema and seed data
	db.InitializeTestDB(t, database)
	db.SeedTestData(t, database)

	return database, cleanup
}

func TestGetProducts(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	// Create server with test database
	srv := New(database)

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/api/products/", nil)
	w := httptest.NewRecorder()

	// Serve the request
	srv.Handler().ServeHTTP(w, req)

	// Check response
	require.Equal(t, http.StatusOK, w.Code)

	// Verify response body contains products
	var products []db.Product
	err := json.NewDecoder(w.Body).Decode(&products)
	require.NoError(t, err)

	// We should have 2 test products
	assert.Len(t, products, 2)

	// Verify first product
	if len(products) > 0 {
		assert.Equal(t, int64(1), products[0].ID)
		assert.Equal(t, "Test Product 1", products[0].Name)
		assert.Equal(t, 19.99, products[0].Price)
	}
}
