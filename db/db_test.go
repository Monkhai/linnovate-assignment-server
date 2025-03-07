package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Fixed time for testing to ensure deterministic results

func setupEmptyDB(t *testing.T) (*DB, func(), context.Context) {
	t.Helper()
	ctx := context.Background()

	db, err := NewTestDB(ctx)
	require.NoError(t, err)

	cleanup := func() {
		CloseTestDB(ctx, db)
	}

	return db, cleanup, ctx
}

func TestDB_Connection(t *testing.T) {
	db, cleanup, ctx := setupEmptyDB(t)
	defer cleanup()
	err := db.pool.Ping(ctx)
	require.NoError(t, err)
}

func TestGetProducts(t *testing.T) {
	db, cleanup, ctx := setupEmptyDB(t)
	defer cleanup()

	testProducts := []Product{
		{ID: 1, Name: "Test Product 1", Price: 19.99, Image: "https://via.placeholder.com/150", Description: "Test Description 1"},
		{ID: 2, Name: "Test Product 2", Price: 29.99, Image: "https://via.placeholder.com/150", Description: "Test Description 2"},
		{ID: 3, Name: "Test Product 3", Price: 39.99, Image: "https://via.placeholder.com/150", Description: "Test Description 3"},
	}

	err := PopulateTestData(ctx, db, "products", testProducts)
	require.NoError(t, err)

	products, err := db.GetProducts(ctx)
	require.NoError(t, err)

	require.Len(t, products, len(testProducts))
	for i, p := range products {
		tp := testProducts[i]
		validateProduct(t, p, tp)
	}
}

func TestGetProduct(t *testing.T) {
	db, cleanup, ctx := setupEmptyDB(t)
	defer cleanup()

	testProducts := []Product{
		{ID: 1, Name: "Test Product 1", Price: 19.99, Image: "https://via.placeholder.com/150", Description: "Test Description 1"},
		{ID: 2, Name: "Test Product 2", Price: 29.99, Image: "https://via.placeholder.com/150", Description: "Test Description 2"},
		{ID: 3, Name: "Test Product 3", Price: 39.99, Image: "https://via.placeholder.com/150", Description: "Test Description 3"},
	}

	err := PopulateTestData(ctx, db, "products", testProducts)
	require.NoError(t, err)

	for _, tp := range testProducts {
		p, err := db.GetProduct(ctx, tp.ID)
		require.NoError(t, err)
		validateProduct(t, p, tp)
	}
}

func TestPostReview(t *testing.T) {
	db, cleanup, ctx := setupEmptyDB(t)
	defer cleanup()

	testProducts := []Product{
		{ID: 1, Name: "Test Product 1", Price: 19.99, Image: "https://via.placeholder.com/150", Description: "Test Description 1"},
		{ID: 2, Name: "Test Product 2", Price: 29.99, Image: "https://via.placeholder.com/150", Description: "Test Description 2"},
		{ID: 3, Name: "Test Product 3", Price: 39.99, Image: "https://via.placeholder.com/150", Description: "Test Description 3"},
	}

	err := PopulateTestData(ctx, db, "products", testProducts)
	require.NoError(t, err)

	testReviews := []Review{
		{ID: 1, UserId: "1", ProductID: 1, ReviewTitle: "Title 1", ReviewContent: "Content 1", Stars: 1},
		{ID: 2, UserId: "1", ProductID: 2, ReviewTitle: "Title 2", ReviewContent: "Content 2", Stars: 2},
		{ID: 3, UserId: "1", ProductID: 3, ReviewTitle: "Title 3", ReviewContent: "Content 3", Stars: 3},
	}

	for _, tr := range testReviews {
		r, err := db.PostReview(ctx, ClientReview{ProductID: tr.ProductID, ReviewTitle: tr.ReviewTitle, ReviewContent: tr.ReviewContent, Stars: tr.Stars}, "1")
		require.NoError(t, err)
		validateReview(t, r, tr)
	}
}

func TestGetProductReviews(t *testing.T) {
	db, cleanup, ctx := setupEmptyDB(t)
	defer cleanup()

	testProducts := []Product{
		{ID: 1, Name: "Test Product 1", Price: 19.99, Image: "https://via.placeholder.com/150", Description: "Test Description 1"},
		{ID: 2, Name: "Test Product 2", Price: 29.99, Image: "https://via.placeholder.com/150", Description: "Test Description 2"},
		{ID: 3, Name: "Test Product 3", Price: 39.99, Image: "https://via.placeholder.com/150", Description: "Test Description 3"},
	}
	err := PopulateTestData(ctx, db, "products", testProducts)
	require.NoError(t, err)

	testReviews := []Review{
		{ID: 1, UserId: "1", ProductID: 1, ReviewTitle: "Title 1", ReviewContent: "Content 1", Stars: 1},
		{ID: 2, UserId: "1", ProductID: 1, ReviewTitle: "Title 2", ReviewContent: "Content 2", Stars: 2},
		{ID: 3, UserId: "1", ProductID: 2, ReviewTitle: "Title 3", ReviewContent: "Content 3", Stars: 3},
		{ID: 4, UserId: "1", ProductID: 2, ReviewTitle: "Title 4", ReviewContent: "Content 4", Stars: 1},
		{ID: 5, UserId: "1", ProductID: 3, ReviewTitle: "Title 5", ReviewContent: "Content 5", Stars: 2},
		{ID: 6, UserId: "1", ProductID: 3, ReviewTitle: "Title 6", ReviewContent: "Content 6", Stars: 3},
	}
	err = PopulateTestData(ctx, db, "reviews", testReviews)
	require.NoError(t, err)

	tests := []struct {
		productId int64
		expected  []Review
	}{
		{
			productId: 1,
			expected: []Review{
				{ID: 1, UserId: "1", ProductID: 1, ReviewTitle: "Title 1", ReviewContent: "Content 1", Stars: 1},
				{ID: 2, UserId: "1", ProductID: 1, ReviewTitle: "Title 2", ReviewContent: "Content 2", Stars: 2},
			},
		},
		{
			productId: 2,
			expected: []Review{
				{ID: 3, UserId: "1", ProductID: 2, ReviewTitle: "Title 3", ReviewContent: "Content 3", Stars: 3},
				{ID: 4, UserId: "1", ProductID: 2, ReviewTitle: "Title 4", ReviewContent: "Content 4", Stars: 1},
			},
		},
		{
			productId: 3,
			expected: []Review{
				{ID: 5, UserId: "1", ProductID: 3, ReviewTitle: "Title 5", ReviewContent: "Content 5", Stars: 2},
				{ID: 6, UserId: "1", ProductID: 3, ReviewTitle: "Title 6", ReviewContent: "Content 6", Stars: 3},
			},
		},
	}

	for _, tt := range tests {
		reviews, err := db.GetProductReviews(ctx, tt.productId)
		require.NoError(t, err)
		assert.Len(t, reviews, len(tt.expected))
		for i, r := range reviews {
			tr := tt.expected[i]
			validateReview(t, r, tr)
		}
	}

}

// ===========================================
// =================HELPERS===================
// ===========================================

func validateProduct(t *testing.T, p, tp Product) {
	assert.Equal(t, tp.ID, p.ID)
	assert.Equal(t, tp.Name, p.Name)
	assert.Equal(t, tp.Price, p.Price)
	assert.Equal(t, tp.Image, p.Image)
	assert.Equal(t, tp.Description, p.Description)
}

func validateReview(t *testing.T, r, tr Review) {
	assert.Equal(t, r.ID, tr.ID)
	assert.Equal(t, r.ProductID, tr.ProductID)
	assert.Equal(t, r.UserId, tr.UserId)
	assert.Equal(t, r.ReviewTitle, tr.ReviewTitle)
	assert.Equal(t, r.ReviewContent, tr.ReviewContent)
	assert.Equal(t, r.Stars, tr.Stars)

}
