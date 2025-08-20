package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultPostgreSQLConfig(t *testing.T) {
	config := DefaultPostgreSQLConfig()

	assert.Equal(t, "postgres", config.Driver)
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 5432, config.Port)
	assert.Equal(t, "postgres", config.Username)
	assert.Equal(t, "password", config.Password)
	assert.Equal(t, "myapp", config.DatabaseName)
	assert.Equal(t, "disable", config.SSLMode)
	assert.Equal(t, 25, config.MaxOpenConns)
	assert.Equal(t, 5, config.MaxIdleConns)
	assert.Equal(t, 5*time.Minute, config.ConnMaxLifetime)
	assert.Equal(t, 5*time.Minute, config.ConnMaxIdleTime)
}

func TestDefaultMySQLConfig(t *testing.T) {
	config := DefaultMySQLConfig()

	assert.Equal(t, "mysql", config.Driver)
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 3306, config.Port)
	assert.Equal(t, "root", config.Username)
	assert.Equal(t, "password", config.Password)
	assert.Equal(t, "myapp", config.DatabaseName)
	assert.Equal(t, 25, config.MaxOpenConns)
	assert.Equal(t, 5, config.MaxIdleConns)
	assert.Equal(t, 5*time.Minute, config.ConnMaxLifetime)
	assert.Equal(t, 5*time.Minute, config.ConnMaxIdleTime)
}

func TestBuildDSN_PostgreSQL(t *testing.T) {
	config := &DatabaseConfig{
		Driver:       "postgres",
		Host:         "localhost",
		Port:         5432,
		Username:     "testuser",
		Password:     "testpass",
		DatabaseName: "testdb",
		SSLMode:      "disable",
	}

	dsn, err := buildDSN(config)
	assert.NoError(t, err)

	expected := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
	assert.Equal(t, expected, dsn)
}

func TestBuildDSN_MySQL(t *testing.T) {
	config := &DatabaseConfig{
		Driver:       "mysql",
		Host:         "localhost",
		Port:         3306,
		Username:     "testuser",
		Password:     "testpass",
		DatabaseName: "testdb",
	}

	dsn, err := buildDSN(config)
	assert.NoError(t, err)

	expected := "testuser:testpass@tcp(localhost:3306)/testdb?parseTime=true"
	assert.Equal(t, expected, dsn)
}

func TestBuildDSN_UnsupportedDriver(t *testing.T) {
	config := &DatabaseConfig{
		Driver: "unsupported",
	}

	dsn, err := buildDSN(config)
	assert.Error(t, err)
	assert.Empty(t, dsn)
	assert.Contains(t, err.Error(), "unsupported database driver")
}

func TestNewDatabase_InvalidDriver(t *testing.T) {
	config := &DatabaseConfig{
		Driver: "invalid",
	}

	db, err := NewDatabase(config)
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "unsupported database driver")
}

// Note: The following tests would require actual database connections
// In a real scenario, you might use test containers or mock databases

func TestDatabaseConfig_Validation(t *testing.T) {
	tests := []struct {
		name     string
		config   *DatabaseConfig
		hasError bool
	}{
		{
			name:     "Valid PostgreSQL config",
			config:   DefaultPostgreSQLConfig(),
			hasError: false,
		},
		{
			name:     "Valid MySQL config",
			config:   DefaultMySQLConfig(),
			hasError: false,
		},
		{
			name: "Invalid driver",
			config: &DatabaseConfig{
				Driver: "invalid",
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := buildDSN(tt.config)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
