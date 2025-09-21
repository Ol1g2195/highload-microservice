package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"

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
	// Read migrations file
	migrationsPath := filepath.Join("internal", "database", "migrations.sql")
	migrations, err := ioutil.ReadFile(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to read migrations file: %w", err)
	}

	// Execute migrations
	if _, err := db.Exec(string(migrations)); err != nil {
		return fmt.Errorf("failed to execute migrations: %w", err)
	}

	return nil
}

