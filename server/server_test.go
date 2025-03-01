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

	return database, cleanup
}

func TestGetProducts(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	srv := New(database)

	testProducts := []db.Product{
		{ID: 1, Name: "Test Product 1", Price: 19.99},
		{ID: 2, Name: "Test Product 2", Price: 29.99},
		{ID: 3, Name: "Test Product 3", Price: 39.99},
	}

	ctx := context.Background()
	err := db.PopulateTestData(ctx, database, "products", testProducts)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/products/", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var products []db.Product
	err = json.NewDecoder(w.Body).Decode(&products)
	require.NoError(t, err)

	require.Equal(t, len(testProducts), len(products))

	for i, p := range products {
		validateProduct(t, p, testProducts[i])
	}
}

// ===========================================
// =================HELPERS===================
// ===========================================

func validateProduct(t *testing.T, p, tp db.Product) {
	assert.Equal(t, tp.ID, p.ID)
	assert.Equal(t, tp.Name, p.Name)
	assert.Equal(t, tp.Price, p.Price)
}

// func validateReview(t *testing.T, r, tr db.Review) {
// 	assert.Equal(t, r.ID, tr.ID)
// 	assert.Equal(t, r.ProductID, tr.ProductID)
// 	assert.Equal(t, r.UserId, tr.UserId)
// 	assert.Equal(t, r.ReviewTitle, tr.ReviewTitle)
// 	assert.Equal(t, r.ReviewContent, tr.ReviewContent)
// 	assert.Equal(t, r.Stars, tr.Stars)
// }
