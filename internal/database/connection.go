package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Driver          string
	Host            string
	Port            int
	Username        string
	Password        string
	DatabaseName    string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// Database wraps the sql.DB connection with additional functionality
type Database struct {
	DB     *sql.DB
	Config *DatabaseConfig
}

// NewDatabase creates a new database connection with the provided configuration
func NewDatabase(config *DatabaseConfig) (*Database, error) {
	dsn, err := buildDSN(config)
	if err != nil {
		return nil, fmt.Errorf("failed to build DSN: %w", err)
	}

	db, err := sql.Open(config.Driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	database := &Database{
		DB:     db,
		Config: config,
	}

	// Test the connection
	if err := database.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Successfully connected to %s database", config.Driver)
	return database, nil
}

// buildDSN constructs the data source name based on the database driver
func buildDSN(config *DatabaseConfig) (string, error) {
	switch config.Driver {
	case "postgres":
		return fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			config.Host,
			config.Port,
			config.Username,
			config.Password,
			config.DatabaseName,
			config.SSLMode,
		), nil
	case "mysql":
		return fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?parseTime=true",
			config.Username,
			config.Password,
			config.Host,
			config.Port,
			config.DatabaseName,
		), nil
	default:
		return "", fmt.Errorf("unsupported database driver: %s", config.Driver)
	}
}

// Ping checks if the database connection is alive
func (d *Database) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return d.DB.PingContext(ctx)
}

// HealthCheck performs a comprehensive health check of the database
func (d *Database) HealthCheck() error {
	// Check if connection is alive
	if err := d.Ping(); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	// Check connection pool stats
	stats := d.DB.Stats()
	if stats.OpenConnections == 0 {
		return fmt.Errorf("no open connections available")
	}

	// Perform a simple query to ensure database is responsive
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var result int
	query := "SELECT 1"
	if d.Config.Driver == "mysql" {
		query = "SELECT 1"
	} else if d.Config.Driver == "postgres" {
		query = "SELECT 1"
	}

	err := d.DB.QueryRowContext(ctx, query).Scan(&result)
	if err != nil {
		return fmt.Errorf("health check query failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected health check result: %d", result)
	}

	return nil
}

// Close closes the database connection
func (d *Database) Close() error {
	if d.DB != nil {
		log.Println("Closing database connection")
		return d.DB.Close()
	}
	return nil
}

// GetStats returns database connection pool statistics
func (d *Database) GetStats() sql.DBStats {
	return d.DB.Stats()
}

// DefaultPostgreSQLConfig returns a default configuration for PostgreSQL
func DefaultPostgreSQLConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Driver:          "postgres",
		Host:            "localhost",
		Port:            5432,
		Username:        "postgres",
		Password:        "password",
		DatabaseName:    "myapp",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}
}

// DefaultMySQLConfig returns a default configuration for MySQL
func DefaultMySQLConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Driver:          "mysql",
		Host:            "localhost",
		Port:            3306,
		Username:        "root",
		Password:        "password",
		DatabaseName:    "myapp",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}
}
