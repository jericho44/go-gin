# Repository Package

This package contains the repository pattern implementation for data access operations in the Gin Golang application.

## Structure

```
internal/repository/
├── README.md           # This documentation
├── base.go            # Base repository interface and implementation
├── user.go            # User repository interface and implementation
├── user_test.go       # Unit tests for user repository
└── user_example.go    # Usage examples for user repository
```

## Overview

The repository package provides a clean abstraction layer for database operations, following the Repository pattern. It separates data access logic from business logic, making the codebase more maintainable and testable.

## Key Components

### Base Repository (`base.go`)

- **Repository Interface**: Defines common functionality for all repositories
- **BaseRepository**: Provides shared functionality like database connection access and transaction management

### User Repository (`user.go`)

- **UserRepository Interface**: Defines all user-related data access operations
- **userRepository Implementation**: Concrete implementation with full CRUD operations

## Features

- **Multi-database Support**: Works with PostgreSQL, MySQL, and SQLite
- **Database-agnostic Queries**: Automatic parameter placeholder handling for different database drivers
- **Transaction Support**: Built-in transaction capabilities through base repository
- **Comprehensive Error Handling**: Proper error messages and handling for all operations
- **Full CRUD Operations**: Create, Read, Update, Delete operations for User model
- **Additional Operations**: Exists checks, counting, and paginated listing

## Usage

### Basic Usage

```go
import (
    "gin-golang-app/internal/database"
    "gin-golang-app/internal/repository"
)

// Create database connection
config := database.DefaultPostgreSQLConfig()
db, err := database.NewDatabase(config)
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Create user repository
userRepo := repository.NewUserRepository(db)

// Use the repository
ctx := context.Background()
user := &models.User{Name: "John Doe", Email: "john@example.com"}
err = userRepo.Create(ctx, user)
```

### Available Operations

- `Create(ctx, user)` - Create a new user
- `GetByID(ctx, id)` - Get user by ID
- `GetByEmail(ctx, email)` - Get user by email
- `Update(ctx, user)` - Update existing user
- `Delete(ctx, id)` - Delete user by ID
- `List(ctx, limit, offset)` - Get paginated list of users
- `Count(ctx)` - Get total user count
- `Exists(ctx, id)` - Check if user exists by ID
- `ExistsByEmail(ctx, email)` - Check if user exists by email

### Transaction Support

```go
// Begin transaction
tx, err := userRepo.BeginTx(ctx)
if err != nil {
    log.Fatal(err)
}

// Perform operations...
err = userRepo.Create(ctx, user1)
if err != nil {
    tx.Rollback()
    return err
}

// Commit transaction
err = tx.Commit()
```

## Testing

Run the repository tests:

```bash
go test ./internal/repository -v
```

## Examples

See `user_example.go` for comprehensive usage examples including:

- Basic CRUD operations
- Transaction handling
- Error handling patterns

## Database Compatibility

The repository automatically handles different database drivers:

- **PostgreSQL**: Uses `$1, $2, ...` parameter placeholders
- **MySQL**: Uses `?, ?, ...` parameter placeholders
- **SQLite**: Uses `?, ?, ...` parameter placeholders

## Error Handling

All repository methods return descriptive errors:

- "not found" errors for missing records
- Database connection errors
- Constraint violation errors
- Transaction errors

## Future Extensions

To add new repositories:

1. Create a new interface extending the base `Repository` interface
2. Implement the concrete repository struct embedding `*BaseRepository`
3. Add appropriate tests and examples
4. Follow the same patterns established in the user repository
