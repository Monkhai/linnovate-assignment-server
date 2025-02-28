package db

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type DB struct {
	client *pgxpool.Pool
	ctx    context.Context

	// Container fields, only used when in test mode
	container testcontainers.Container
	testMode  bool
}

// Options for DB creation
type Options struct {
	// Connection options
	ConnectionString string

	// Test container options
	TestMode bool
	Database string
	User     string
	Password string
}

// New creates a regular DB connection using the provided connection string
func New(ctx context.Context, client *pgxpool.Pool) *DB {
	return &DB{ctx: ctx, client: client}
}

// NewTest creates a new DB with more control over configuration
// If testMode is true, it will create a test container
func NewTest(ctx context.Context) (*DB, error) {
	opts := Options{
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
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	// Test the connection
	if err := client.Ping(ctx); err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return &DB{
		ctx:       ctx,
		client:    client,
		container: container,
		testMode:  true,
	}, nil
}

// Close closes the DB connection and terminates the container if in test mode
func (db *DB) Close(ctx context.Context) {
	if db.client != nil {
		db.client.Close()
	}

	if db.testMode && db.container != nil {
		db.container.Terminate(ctx)
	}
}

func (db *DB) populateTestData(ctx context.Context, tableName string, data interface{}) error {
	if !db.testMode {
		return fmt.Errorf("populate only works in test mode")
	}

	val := reflect.ValueOf(data)
	if val.Kind() != reflect.Slice {
		return fmt.Errorf("expected slice, got %T", data)
	}
	if val.Len() == 0 {
		return nil
	}

	tx, err := db.client.Begin(ctx)
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

func (db *DB) GetProducts() ([]Product, error) {
	rows, err := db.client.Query(db.ctx, "SELECT id, name, price, created_at FROM products ORDER BY $1", "id")
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
