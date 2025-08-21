# Integration Tests

This directory contains integration tests for the Gin Golang application. These tests verify the complete request-response cycles, middleware functionality, and database operations.

## Prerequisites

Before running integration tests, ensure you have:

1. **PostgreSQL Database**: A PostgreSQL instance running locally or accessible remotely
2. **Go Environment**: Go 1.21 or later installed
3. **Dependencies**: All Go modules installed (`go mod download`)

## Database Setup

The integration tests require a PostgreSQL database. The tests will:

1. Create a test database (`gin_app_test` by default)
2. Create necessary tables and triggers
3. Clean up data between tests
4. Drop the test database after all tests complete

### Default Database Configuration

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=
DB_NAME=gin_app_test
```

### Custom Database Configuration

You can override the default configuration by setting environment variables:

```bash
export DB_HOST=your-db-host
export DB_PORT=5432
export DB_USER=your-username
export DB_PASSWORD=your-password
export DB_NAME=your-test-db-name
```

## Running Tests

### Run All Integration Tests

```bash
# From the project root
go test ./tests/integration/api/... -v

# Or from the integration test directory
cd tests/integration/api
go test -v
```

### Run Specific Test Suites

```bash
# Run only health endpoint tests
go test ./tests/integration/api -run TestHealthIntegrationSuite -v

# Run only user endpoint tests
go test ./tests/integration/api -run TestUserIntegrationSuite -v

# Run only authentication tests
go test ./tests/integration/api -run TestAuthIntegrationSuite -v

# Run only database tests
go test ./tests/integration/api -run TestDatabaseIntegrationSuite -v
```

### Run Specific Test Cases

```bash
# Run a specific test case
go test ./tests/integration/api -run TestUserIntegrationSuite/TestCreateUser -v

# Run tests matching a pattern
go test ./tests/integration/api -run ".*User.*" -v
```

## Test Structure

### Test Suites

1. **HealthIntegrationTestSuite** (`health_test.go`)

   - Tests health check endpoints (`/health`, `/ready`, `/live`)
   - Verifies endpoints work without authentication
   - Validates response structure and headers

2. **UserIntegrationTestSuite** (`user_test.go`)

   - Tests all user CRUD operations
   - Validates request/response formats
   - Tests pagination functionality
   - Verifies database operations

3. **AuthIntegrationTestSuite** (`auth_test.go`)

   - Tests JWT authentication
   - Validates middleware functionality (CORS, logging, request ID)
   - Tests protected endpoint access control

4. **DatabaseIntegrationTestSuite** (`database_test.go`)
   - Tests database connections and transactions
   - Validates database constraints and triggers
   - Tests cleanup and sequence reset functionality

### Test Utilities

- **`setup_test.go`**: Base test suite with database setup and teardown
- **`fixtures.go`**: Test data fixtures and database helper functions
- **`helpers.go`**: HTTP client utilities and JWT token generation
- **`main_test.go`**: Test runner with environment setup

## Test Data Management

### Fixtures

Test fixtures are defined in `fixtures.go` and provide:

- Predefined user data for consistent testing
- Invalid request data for validation testing
- Helper functions for seeding and querying test data

### Database Cleanup

The test suite automatically:

- Cleans up all data between tests
- Resets database sequences
- Maintains test isolation

### Seeding Data

```go
// Seed a single user
user := suite.SeedUser(UserFixtures.ValidUser1)

// Seed multiple users
users := suite.SeedUsers(UserFixtures.ValidUser1, UserFixtures.ValidUser2)

// Check if user exists
exists := suite.UserExistsInDB(userID)
```

## Authentication Testing

### JWT Token Generation

```go
// Generate a test JWT token
token, err := GenerateJWTToken("test-secret-key", userID, email)

// Create authenticated HTTP client
client, token, err := suite.GetAuthenticatedHTTPClient(userID, email)

// Make authenticated requests
resp, err := client.GET("/api/v1/users", AuthHeaders(token))
```

## Environment Variables

The tests use the following environment variables:

| Variable      | Default           | Description              |
| ------------- | ----------------- | ------------------------ |
| `APP_ENV`     | `test`            | Application environment  |
| `DB_HOST`     | `localhost`       | Database host            |
| `DB_PORT`     | `5432`            | Database port            |
| `DB_USER`     | `postgres`        | Database username        |
| `DB_PASSWORD` | ``                | Database password        |
| `DB_NAME`     | `gin_app_test`    | Test database name       |
| `JWT_SECRET`  | `test-secret-key` | JWT signing secret       |
| `SERVER_HOST` | `localhost`       | Server host              |
| `SERVER_PORT` | `0`               | Server port (0 = random) |

## Troubleshooting

### Database Connection Issues

1. **PostgreSQL not running**: Ensure PostgreSQL is running and accessible
2. **Permission denied**: Check database user permissions
3. **Database doesn't exist**: Tests will create the database automatically

### Test Failures

1. **Port conflicts**: Tests use random ports to avoid conflicts
2. **Data isolation**: Each test runs with a clean database state
3. **Timing issues**: Tests include proper cleanup and setup

### Common Commands

```bash
# Check PostgreSQL status
pg_ctl status

# Start PostgreSQL (macOS with Homebrew)
brew services start postgresql

# Start PostgreSQL (Linux)
sudo systemctl start postgresql

# Connect to PostgreSQL
psql -U postgres -h localhost
```

## CI/CD Integration

For continuous integration, ensure:

1. PostgreSQL service is available
2. Environment variables are set
3. Database permissions are configured
4. Tests run with appropriate timeouts

Example GitHub Actions configuration:

```yaml
services:
  postgres:
    image: postgres:13
    env:
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: gin_app_test
    options: >-
      --health-cmd pg_isready
      --health-interval 10s
      --health-timeout 5s
      --health-retries 5

steps:
  - name: Run integration tests
    run: go test ./tests/integration/api/... -v
    env:
      DB_HOST: localhost
      DB_PORT: 5432
      DB_USER: postgres
      DB_PASSWORD: postgres
      DB_NAME: gin_app_test
```
