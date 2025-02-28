package db

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// testContainer tracks a test container for cleanup
type testContainer struct {
	container testcontainers.Container
}

// testContainers keeps track of all containers created for tests
var testContainers = make(map[*DB]*testContainer)

// NewTestDB creates a new DB backed by a PostgreSQL test container
func NewTestDB(ctx context.Context) (*DB, error) {
	opts := struct {
		Database string
		User     string
		Password string
	}{
		Database: "test_db",
		User:     "test_user",
		Password: "test_password",
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

	// Set up cleanup if something fails
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

	// Wait a bit to ensure Postgres is ready for connections
	time.Sleep(time.Second * 2)

	client, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	// Test the connection
	if err := client.Ping(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	// Apply default migrations for testing
	if err := applyDefaultMigrations(ctx, client); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	// Create a regular DB instance
	db := New(ctx, client)

	// Track the container for later cleanup
	testContainers[db] = &testContainer{container: container}

	return db, nil
}

// CloseTestDB properly cleans up a test database including its container
func CloseTestDB(ctx context.Context, db *DB) {
	if db == nil {
		return
	}

	// First close the database connection
	if db.client != nil {
		db.client.Close()
	}

	// Then terminate the container if this was a test DB
	if tc, exists := testContainers[db]; exists && tc.container != nil {
		tc.container.Terminate(ctx)
		delete(testContainers, db)
	}
}

// PopulateTestData populates a table with test data for testing purposes
func PopulateTestData(db *DB, tableName string, data interface{}) error {
	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Slice {
		return fmt.Errorf("expected slice, got %T", data)
	}
	if val.Len() == 0 {
		return nil
	}

	tx, err := db.client.Begin(db.ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(db.ctx)

	elem := val.Index(0).Interface()
	elemType := reflect.TypeOf(elem)
	if elemType.Kind() != reflect.Struct {
		return fmt.Errorf("expected slice of structs, got slice of %s", elemType.Kind())
	}

	var fields []string
	for i := 0; i < elemType.NumField(); i++ {
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
	args := make([]interface{}, 0, val.Len()*len(fields))

	for i := 0; i < val.Len(); i++ {
		paramOffset := i * len(fields)
		placeholders := make([]string, len(fields))
		for j := 0; j < len(fields); j++ {
			placeholders[j] = fmt.Sprintf("$%d", paramOffset+j+1)

			// Extract field value
			fieldValue := val.Index(i).Field(j).Interface()
			args = append(args, fieldValue)
		}
		placeholderGroups[i] = fmt.Sprintf("(%s)", strings.Join(placeholders, ", "))
	}

	sb.WriteString(strings.Join(placeholderGroups, ", "))

	_, err = tx.Exec(db.ctx, sb.String(), args...)
	if err != nil {
		return fmt.Errorf("failed to insert data: %w", err)
	}

	err = tx.Commit(db.ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
