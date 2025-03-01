package server

import (
	"catalogapi/config"
	"catalogapi/db"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

// Custom context key type to avoid collisions
type contextKey string

// Define key constants
const userIDKey contextKey = "userID"

// Server represents the HTTP server and its dependencies
type Server struct {
	router http.Handler
	db     *db.DB
	auth   *auth.Client
}

// New creates a new server instance with all required dependencies
func New(database *db.DB) *Server {
	auth, err := newAuthClient(context.Background())
	if err != nil {
		log.Fatalf("Failed to create auth client: %v", err)
	}
	s := &Server{
		db:   database,
		auth: auth,
	}
	s.setupRoutes()
	return s
}

func authMiddleware(auth *auth.Client, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := authorize(auth, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next(w, r.WithContext(ctx))
	}
}

// corsMiddleware adds CORS headers to allow requests from localhost:3000
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for ALL requests
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight OPTIONS requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// setupRoutes configures all the API routes
func (s *Server) setupRoutes() {
	mux := http.NewServeMux()

	// Add routes with and without trailing slash
	mux.HandleFunc("GET /api/products", s.getProducts)
	mux.HandleFunc("POST /api/reviews", authMiddleware(s.auth, s.postReview))
	mux.HandleFunc("GET /api/products/{id}/reviews", s.getProductReviews)

	// Add CORS middleware to the router
	s.router = corsMiddleware(mux)
}

// Handler returns the HTTP handler for the server
func (s *Server) Handler() http.Handler {
	// Make extra sure that CORS middleware is applied
	return corsMiddleware(s.router)
}

func (s *Server) getProducts(w http.ResponseWriter, r *http.Request) {
	products, err := s.db.GetProducts(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(products)
}

func (s *Server) postReview(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value(userIDKey).(string)
	if !ok {
		http.Error(w, "user not found in context", http.StatusUnauthorized)
		return
	}

	var clientReview db.ClientReview
	err := json.NewDecoder(r.Body).Decode(&clientReview)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	review, err := s.db.PostReview(r.Context(), clientReview, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	safeReview := db.SafeReview{
		ID:            review.ID,
		ProductID:     clientReview.ProductID,
		ReviewTitle:   clientReview.ReviewTitle,
		ReviewContent: clientReview.ReviewContent,
		Stars:         clientReview.Stars,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(safeReview)
}

func (s *Server) getProductReviews(w http.ResponseWriter, r *http.Request) {
	productId := r.PathValue("id")
	if productId == "" {
		http.Error(w, "product id is required", http.StatusBadRequest)
		return
	}
	productIdInt, err := strconv.ParseInt(productId, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	reviews, err := s.db.GetProductReviews(r.Context(), productIdInt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	safeReviews := make([]db.SafeReview, len(reviews))
	for i, review := range reviews {
		safeReviews[i] = db.SafeReview{
			ID:            review.ID,
			ProductID:     review.ProductID,
			ReviewTitle:   review.ReviewTitle,
			ReviewContent: review.ReviewContent,
			Stars:         review.Stars,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(safeReviews)
}

// ===========================================
// =================HELPERS===================
// ===========================================

func newAuthClient(ctx context.Context) (*auth.Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	opt := option.WithCredentialsFile(cfg.Firebase.CredentialsFile)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, err
	}

	auth, err := app.Auth(ctx)
	if err != nil {
		return nil, err
	}

	return auth, nil
}

func authorize(auth *auth.Client, r *http.Request) (string, error) {
	token := r.Header.Get("Authorization")
	if token == "" {
		return "", fmt.Errorf("missing Authorization header")
	}
	ctx := context.Background()
	u, err := auth.VerifyIDToken(ctx, token)
	if err != nil {
		return "", fmt.Errorf("invalid Authorization header: %w", err)
	}

	return u.UID, nil
}
