package config

import (
	"log"
	"os"
	"testing"
)

func TestLoadProductionConfig(t *testing.T) {
	os.Setenv("APP_ENV", "production")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Environment != "production" {
		t.Errorf("Expected environment to be 'production', got '%s'", cfg.Environment)
	}

	log.Printf("Config: %+v", cfg.Database)
}
