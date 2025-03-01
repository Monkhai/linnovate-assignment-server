package db

import (
	"catalogapi/config"
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// InitializeTestDB initializes the database schema for testing
func InitializeTestDB(t *testing.T, db *DB) {
	ctx := context.Background()

	// Create products table if it doesn't exist
	_, err := db.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS products (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			price FLOAT NOT NULL,
			image TEXT NOT NULL,
			description TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err)

	// Create reviews table if it doesn't exist
	_, err = db.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS reviews (
			id SERIAL PRIMARY KEY,
			user_id TEXT NOT NULL,
			product_id INTEGER NOT NULL REFERENCES products(id),
			review_title TEXT,
			review_content TEXT,
			stars INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err)
}

// CleanupTestDB removes test data from the database
func CleanupTestDB(t *testing.T, db *DB) {
	ctx := context.Background()
	_, err := db.pool.Exec(ctx, "TRUNCATE products CASCADE, reviews CASCADE")
	require.NoError(t, err)
}

type testContainer struct {
	container testcontainers.Container
}

var testContainers = make(map[*DB]*testContainer)

func NewTestDB(ctx context.Context) (*DB, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	opts := struct {
		Database string
		User     string
		Password string
	}{
		Database: cfg.Database.Name,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
	}

	req := testcontainers.ContainerRequest{
		Image:        "postgres:13",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     opts.User,
			"POSTGRES_PASSWORD": opts.Password,
			"POSTGRES_DB":       opts.Database,
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(time.Second * 30),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	containerCleanup := func() {
		container.Terminate(ctx)
	}
	defer func() {
		if err != nil {
			containerCleanup()
		}
	}()

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, fmt.Errorf("failed to get container port: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		opts.User, opts.Password, host, port.Port(), opts.Database)

	time.Sleep(time.Second * 2)

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	if err := applyDefaultMigrations(ctx, pool); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	db := &DB{pool: pool}

	testContainers[db] = &testContainer{container: container}

	return db, nil
}

func CloseTestDB(ctx context.Context, db *DB) {
	if db == nil {
		return
	}

	if db.pool != nil {
		db.pool.Close()
	}

	if tc, exists := testContainers[db]; exists && tc.container != nil {
		tc.container.Terminate(ctx)
		delete(testContainers, db)
	}
}

func PopulateTestData(ctx context.Context, db *DB, tableName string, data any) error {
	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Slice {
		return fmt.Errorf("expected slice, got %T", data)
	}
	if val.Len() == 0 {
		return nil
	}

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	elem := val.Index(0).Interface()
	elemType := reflect.TypeOf(elem)
	if elemType.Kind() != reflect.Struct {
		return fmt.Errorf("expected slice of structs, got slice of %s", elemType.Kind())
	}

	var fields []string
	for i := range elemType.NumField() {
		field := elemType.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			dbTag = strings.ToLower(field.Name)
		}
		fields = append(fields, dbTag)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("INSERT INTO %s (", tableName))
	for i, field := range fields {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(field)
	}
	sb.WriteString(") VALUES ")

	placeholderGroups := make([]string, val.Len())
	args := make([]any, 0, val.Len()*len(fields))

	for i := range val.Len() {
		paramOffset := i * len(fields)
		placeholders := make([]string, len(fields))
		for j := range fields {
			placeholders[j] = fmt.Sprintf("$%d", paramOffset+j+1)

			// Extract field value
			fieldValue := val.Index(i).Field(j).Interface()
			args = append(args, fieldValue)
		}
		placeholderGroups[i] = fmt.Sprintf("(%s)", strings.Join(placeholders, ", "))
	}

	sb.WriteString(strings.Join(placeholderGroups, ", "))

	_, err = tx.Exec(ctx, sb.String(), args...)
	if err != nil {
		return fmt.Errorf("failed to insert data: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
