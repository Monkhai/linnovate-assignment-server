package config

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/joho/godotenv"
)

type Config struct {
	Environment string
	Server      ServerConfig
	Database    DatabaseConfig
	Firebase    FirebaseConfig
}

type ServerConfig struct {
	Port int
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type FirebaseConfig struct {
	CredentialsFile string
}

// DBSecret represents the structure of the database secret in AWS Secrets Manager
type DBSecret struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"dbname"`
}

func Load() (*Config, error) {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
		os.Setenv("APP_ENV", env)
	}
	godotenv.Load(fmt.Sprintf(".env.%s", env))

	// Create basic config
	cfg := &Config{
		Environment: env,
		Server: ServerConfig{
			Port: getEnvAsInt("SERVER_PORT", 8080),
		},
		Firebase: FirebaseConfig{
			CredentialsFile: getEnv("FIREBASE_CREDENTIALS_FILE", "../serviceAccountKey.json"),
		},
	}

	// If in production, load DB config from AWS Secrets Manager
	if env == "production" {
		region := getEnv("AWS_REGION", "eu-north-1")
		secretName := getEnv("AWS_DB_SECRET_NAME", "prod/database/credentials")
		host := getEnv("DB_HOST", "")
		port := getEnvAsInt("DB_PORT", 5432)

		username, password, err := loadDatabaseSecretFromAWS(region, secretName)
		if err != nil {
			// Log the error but continue with environment variables as fallback
			fmt.Printf("Warning: Failed to load DB credentials from AWS: %v\nFalling back to environment variables\n", err)
			return nil, err
		} else {
			// Use the AWS Secrets Manager values
			cfg.Database = DatabaseConfig{
				Host:     host,
				Port:     port,
				User:     username,
				Password: password,
				Name:     "postgres",
				SSLMode:  "require",
			}
		}
	} else {
		// In development/test, load DB config from environment variables
		cfg.Database = loadDatabaseConfigFromEnv(env)
	}

	return cfg, nil
}

func loadDatabaseConfigFromEnv(env string) DatabaseConfig {
	return DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvAsInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		Name:     getEnv("DB_NAME", fmt.Sprintf("postgres_%s", env)),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

func (c *DatabaseConfig) GetDatabaseURL() string {
	password := url.QueryEscape(c.Password)
	url := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, password, c.Host, c.Port, c.Name, c.SSLMode,
	)
	return url
}

func loadDatabaseSecretFromAWS(region, secretName string) (string, string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return "", "", err
	}

	// Create Secrets Manager client
	svc := secretsmanager.NewFromConfig(cfg)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}

	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		return "", "", err
	}
	// Decrypts secret using the associated KMS key.
	var secretString string = *result.SecretString
	//parse from json {hostname: string, password: string} to struct
	var secret struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	err = json.Unmarshal([]byte(secretString), &secret)
	if err != nil {
		return "", "", err
	}

	return secret.Username, secret.Password, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
