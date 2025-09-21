package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"

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
		migrations, err = ioutil.ReadFile(path)
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

