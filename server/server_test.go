package server

import (
	"bytes"
	"catalogapi/db"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*db.DB, func(), context.Context) {
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

	return database, cleanup, ctx
}

func TestGetProducts(t *testing.T) {
	database, cleanup, ctx := setupTestDB(t)
	defer cleanup()
	srv := New(database)

	testProducts := []db.Product{
		{ID: 1, Name: "Test Product 1", Price: 19.99, Image: "https://via.placeholder.com/150", Description: "Test Description 1"},
		{ID: 2, Name: "Test Product 2", Price: 29.99, Image: "https://via.placeholder.com/150", Description: "Test Description 2"},
		{ID: 3, Name: "Test Product 3", Price: 39.99, Image: "https://via.placeholder.com/150", Description: "Test Description 3"},
	}

	err := db.PopulateTestData(ctx, database, "products", testProducts)
	require.NoError(t, err)

	r := httptest.NewRequest(http.MethodGet, "/api/products/", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, r)

	require.Equal(t, http.StatusOK, w.Code)

	var products []db.Product
	err = json.NewDecoder(w.Body).Decode(&products)
	require.NoError(t, err)

	require.Equal(t, len(testProducts), len(products))

	for i, p := range products {
		validateProduct(t, p, testProducts[i])
	}
}

func TestPostReview(t *testing.T) {
	database, cleanup, ctx := setupTestDB(t)
	defer cleanup()
	srv := New(database)

	testProducts := []db.Product{
		{ID: 1, Name: "Test Product 1", Price: 19.99, Image: "https://via.placeholder.com/150", Description: "Test Description 1"},
		{ID: 2, Name: "Test Product 2", Price: 29.99, Image: "https://via.placeholder.com/150", Description: "Test Description 2"},
		{ID: 3, Name: "Test Product 3", Price: 39.99, Image: "https://via.placeholder.com/150", Description: "Test Description 3"},
	}

	err := db.PopulateTestData(ctx, database, "products", testProducts)
	require.NoError(t, err)

	testReviews := []db.Review{
		{ID: 1, UserId: "1", ProductID: 1, ReviewTitle: "Title 1", ReviewContent: "Content 1", Stars: 1},
		{ID: 2, UserId: "1", ProductID: 2, ReviewTitle: "Title 2", ReviewContent: "Content 2", Stars: 2},
		{ID: 3, UserId: "1", ProductID: 3, ReviewTitle: "Title 3", ReviewContent: "Content 3", Stars: 3},
	}

	for _, tr := range testReviews {
		clientReview := db.ClientReview{
			ProductID:     tr.ProductID,
			ReviewTitle:   tr.ReviewTitle,
			ReviewContent: tr.ReviewContent,
			Stars:         tr.Stars,
		}

		jsonData, err := json.Marshal(clientReview)
		require.NoError(t, err)
		body := bytes.NewBuffer(jsonData)

		r := httptest.NewRequest(http.MethodPost, "/api/reviews/", body)
		w := httptest.NewRecorder()
		authCtx := context.WithValue(r.Context(), userIDKey, tr.UserId)
		srv.postReview(w, r.WithContext(authCtx))

		require.Equal(t, http.StatusCreated, w.Code)

		var review db.SafeReview
		err = json.NewDecoder(w.Body).Decode(&review)
		require.NoError(t, err)

		validateSafeReview(t, review, tr)
	}
}

func TestGetProductReviews(t *testing.T) {
	database, cleanup, ctx := setupTestDB(t)
	defer cleanup()
	srv := New(database)

	testProducts := []db.Product{
		{ID: 1, Name: "Test Product 1", Price: 19.99, Image: "https://via.placeholder.com/150", Description: "Test Description 1"},
		{ID: 2, Name: "Test Product 2", Price: 29.99, Image: "https://via.placeholder.com/150", Description: "Test Description 2"},
		{ID: 3, Name: "Test Product 3", Price: 39.99, Image: "https://via.placeholder.com/150", Description: "Test Description 3"},
	}
	err := db.PopulateTestData(ctx, database, "products", testProducts)
	require.NoError(t, err)

	testReviews := []db.Review{
		{ID: 1, UserId: "1", ProductID: 1, ReviewTitle: "Title 1", ReviewContent: "Content 1", Stars: 1},
		{ID: 2, UserId: "1", ProductID: 1, ReviewTitle: "Title 2", ReviewContent: "Content 2", Stars: 2},
		{ID: 3, UserId: "1", ProductID: 2, ReviewTitle: "Title 3", ReviewContent: "Content 3", Stars: 3},
		{ID: 4, UserId: "1", ProductID: 2, ReviewTitle: "Title 4", ReviewContent: "Content 4", Stars: 1},
		{ID: 5, UserId: "1", ProductID: 3, ReviewTitle: "Title 5", ReviewContent: "Content 5", Stars: 2},
		{ID: 6, UserId: "1", ProductID: 3, ReviewTitle: "Title 6", ReviewContent: "Content 6", Stars: 3},
	}
	err = db.PopulateTestData(ctx, database, "reviews", testReviews)
	require.NoError(t, err)

	tests := []struct {
		productId int64
		expected  []db.Review
	}{
		{
			productId: 1,
			expected: []db.Review{
				{ID: 1, UserId: "1", ProductID: 1, ReviewTitle: "Title 1", ReviewContent: "Content 1", Stars: 1},
				{ID: 2, UserId: "1", ProductID: 1, ReviewTitle: "Title 2", ReviewContent: "Content 2", Stars: 2},
			},
		},
		{
			productId: 2,
			expected: []db.Review{
				{ID: 3, UserId: "1", ProductID: 2, ReviewTitle: "Title 3", ReviewContent: "Content 3", Stars: 3},
				{ID: 4, UserId: "1", ProductID: 2, ReviewTitle: "Title 4", ReviewContent: "Content 4", Stars: 1},
			},
		},
		{
			productId: 3,
			expected: []db.Review{
				{ID: 5, UserId: "1", ProductID: 3, ReviewTitle: "Title 5", ReviewContent: "Content 5", Stars: 2},
				{ID: 6, UserId: "1", ProductID: 3, ReviewTitle: "Title 6", ReviewContent: "Content 6", Stars: 3},
			},
		},
	}

	for _, tt := range tests {
		path := fmt.Sprintf("/api/products/%d/reviews/", tt.productId)
		r := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		srv.Handler().ServeHTTP(w, r)

		require.Equal(t, http.StatusOK, w.Code)

		var reviews []db.SafeReview

		err := json.NewDecoder(w.Body).Decode(&reviews)
		require.NoError(t, err)

		assert.Len(t, reviews, len(tt.expected))

		for i, tr := range tt.expected {
			r := reviews[i]
			validateSafeReview(t, r, tr)
		}
	}
}

// ===========================================
// =================HELPERS===================
// ===========================================

func validateProduct(t *testing.T, p, tp db.Product) {
	assert.Equal(t, tp.ID, p.ID)
	assert.Equal(t, tp.Name, p.Name)
	assert.Equal(t, tp.Price, p.Price)
	assert.Equal(t, tp.Image, p.Image)
	assert.Equal(t, tp.Description, p.Description)
}

func validateSafeReview(t *testing.T, r db.SafeReview, tr db.Review) {
	assert.Equal(t, tr.ID, r.ID, "ID")
	assert.Equal(t, tr.ProductID, r.ProductID, "ProductID")
	assert.Equal(t, tr.ReviewTitle, r.ReviewTitle, "ReviewTitle")
	assert.Equal(t, tr.ReviewContent, r.ReviewContent, "ReviewContent")
	assert.Equal(t, tr.Stars, r.Stars, "Stars")
}
