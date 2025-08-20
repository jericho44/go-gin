package database

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

// NewDatabaseFromEnv creates a database connection using environment variables
func NewDatabaseFromEnv() (*Database, error) {
	driver := getEnvOrDefault("DB_DRIVER", "postgres")

	var config *DatabaseConfig

	switch driver {
	case "postgres":
		config = &DatabaseConfig{
			Driver:       "postgres",
			Host:         getEnvOrDefault("DB_HOST", "localhost"),
			Port:         getEnvIntOrDefault("DB_PORT", 5432),
			Username:     getEnvOrDefault("DB_USERNAME", "postgres"),
			Password:     getEnvOrDefault("DB_PASSWORD", "password"),
			DatabaseName: getEnvOrDefault("DB_NAME", "myapp"),
			SSLMode:      getEnvOrDefault("DB_SSLMODE", "disable"),
		}
	case "mysql":
		config = &DatabaseConfig{
			Driver:       "mysql",
			Host:         getEnvOrDefault("DB_HOST", "localhost"),
			Port:         getEnvIntOrDefault("DB_PORT", 3306),
			Username:     getEnvOrDefault("DB_USERNAME", "root"),
			Password:     getEnvOrDefault("DB_PASSWORD", "password"),
			DatabaseName: getEnvOrDefault("DB_NAME", "myapp"),
		}
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}

	// Set connection pool settings from environment or defaults
	config.MaxOpenConns = getEnvIntOrDefault("DB_MAX_OPEN_CONNS", 25)
	config.MaxIdleConns = getEnvIntOrDefault("DB_MAX_IDLE_CONNS", 5)
	config.ConnMaxLifetime = time.Duration(getEnvIntOrDefault("DB_CONN_MAX_LIFETIME_MINUTES", 5)) * time.Minute
	config.ConnMaxIdleTime = time.Duration(getEnvIntOrDefault("DB_CONN_MAX_IDLE_TIME_MINUTES", 5)) * time.Minute

	return NewDatabase(config)
}

// ExampleUsage demonstrates how to use the database connection utilities
func ExampleUsage() {
	// Example 1: Using default PostgreSQL configuration
	fmt.Println("=== Example 1: Default PostgreSQL Configuration ===")
	pgConfig := DefaultPostgreSQLConfig()
	fmt.Printf("PostgreSQL DSN would be: host=%s port=%d user=%s dbname=%s sslmode=%s\n",
		pgConfig.Host, pgConfig.Port, pgConfig.Username, pgConfig.DatabaseName, pgConfig.SSLMode)

	// Example 2: Using default MySQL configuration
	fmt.Println("\n=== Example 2: Default MySQL Configuration ===")
	mysqlConfig := DefaultMySQLConfig()
	fmt.Printf("MySQL DSN would be: %s:***@tcp(%s:%d)/%s?parseTime=true\n",
		mysqlConfig.Username, mysqlConfig.Host, mysqlConfig.Port, mysqlConfig.DatabaseName)

	// Example 3: Custom configuration
	fmt.Println("\n=== Example 3: Custom Configuration ===")
	customConfig := &DatabaseConfig{
		Driver:          "postgres",
		Host:            "db.example.com",
		Port:            5432,
		Username:        "myuser",
		Password:        "mypassword",
		DatabaseName:    "production_db",
		SSLMode:         "require",
		MaxOpenConns:    50,
		MaxIdleConns:    10,
		ConnMaxLifetime: 10 * time.Minute,
		ConnMaxIdleTime: 10 * time.Minute,
	}

	dsn, _ := buildDSN(customConfig)
	fmt.Printf("Custom PostgreSQL DSN: %s\n", dsn)

	// Example 4: Environment-based configuration
	fmt.Println("\n=== Example 4: Environment-based Configuration ===")
	fmt.Println("Set these environment variables:")
	fmt.Println("DB_DRIVER=postgres")
	fmt.Println("DB_HOST=localhost")
	fmt.Println("DB_PORT=5432")
	fmt.Println("DB_USERNAME=myuser")
	fmt.Println("DB_PASSWORD=mypassword")
	fmt.Println("DB_NAME=myapp")
	fmt.Println("DB_SSLMODE=disable")
	fmt.Println("DB_MAX_OPEN_CONNS=25")
	fmt.Println("DB_MAX_IDLE_CONNS=5")

	// Note: Actual database connection would be:
	// db, err := NewDatabaseFromEnv()
	// if err != nil {
	//     log.Fatal("Failed to connect to database:", err)
	// }
	// defer db.Close()
	//
	// // Perform health check
	// if err := db.HealthCheck(); err != nil {
	//     log.Printf("Database health check failed: %v", err)
	// } else {
	//     log.Println("Database is healthy")
	// }
}

// Helper functions for environment variable handling
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Warning: Invalid integer value for %s: %s, using default: %d", key, value, defaultValue)
	}
	return defaultValue
}
