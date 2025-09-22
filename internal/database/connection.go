package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"highload-microservice/internal/config"

	_ "github.com/lib/pq"
)

func NewConnection(cfg config.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return db, nil
}

// RunMigrations executes database migrations
func RunMigrations(db *sql.DB) error {
	// Try different possible paths for migrations file
	possiblePaths := []string{
		"internal/database/migrations.sql",
		"./internal/database/migrations.sql",
		"/app/internal/database/migrations.sql",
		"migrations.sql",
	}

	var migrations []byte
	var err error

	for _, path := range possiblePaths {
		// Validate path to prevent directory traversal
		if !isValidPath(path) {
			continue
		}
		migrations, err = os.ReadFile(path) // #nosec G304 -- Path is validated by isValidPath function
		if err == nil {
			break
		}
	}

	if err != nil {
		return fmt.Errorf("failed to read migrations file from any path: %w", err)
	}

	// Execute migrations
	if _, err := db.Exec(string(migrations)); err != nil {
		return fmt.Errorf("failed to execute migrations: %w", err)
	}

	return nil
}

// isValidPath validates that the path is safe and doesn't contain directory traversal
func isValidPath(path string) bool {
	// Clean the path to resolve any .. or . components
	cleanPath := filepath.Clean(path)

	// Check if the path contains any directory traversal attempts
	if strings.Contains(cleanPath, "..") {
		return false
	}

	// Additional validation: ensure the path is within expected directories
	allowedPrefixes := []string{
		"internal/database/",
		"./internal/database/",
		"/app/internal/database/",
		"migrations.sql",
	}

	for _, prefix := range allowedPrefixes {
		if strings.HasPrefix(cleanPath, prefix) {
			return true
		}
	}

	return false
}
