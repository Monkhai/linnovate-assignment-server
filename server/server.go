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

// setupRoutes configures all the API routes
func (s *Server) setupRoutes() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/products/", s.getProducts)
	mux.HandleFunc("POST /api/reviews/", s.postReview)
	mux.HandleFunc("GET /api/products/{id}/reviews/", s.getProductReviews)

	s.router = mux
}

// Handler returns the HTTP handler for the server
func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) getProducts(w http.ResponseWriter, r *http.Request) {
	products, err := s.db.GetProducts(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(products)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
}

func (s *Server) postReview(w http.ResponseWriter, r *http.Request) {
	userId, err := authorize(s.auth, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var review db.ClientReview
	err = json.NewDecoder(r.Body).Decode(&review)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.db.PostReview(r.Context(), review, userId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(review)
}

func (s *Server) getProductReviews(w http.ResponseWriter, r *http.Request) {
	productId := r.URL.Query().Get("productId")
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
	json.NewEncoder(w).Encode(reviews)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
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
