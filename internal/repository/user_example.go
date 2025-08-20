package repository

import (
	"context"
	"fmt"
	"log"

	"gin-golang-app/internal/database"
	"gin-golang-app/internal/models"
)

// ExampleUserRepository demonstrates how to use the UserRepository
func ExampleUserRepository() {
	// Create database configuration
	config := database.DefaultPostgreSQLConfig()
	config.DatabaseName = "myapp"
	config.Username = "myuser"
	config.Password = "mypassword"

	// Create database connection
	db, err := database.NewDatabase(config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create user repository
	userRepo := NewUserRepository(db)

	ctx := context.Background()

	// Create a new user
	user := &models.User{
		Name:  "John Doe",
		Email: "john.doe@example.com",
	}

	err = userRepo.Create(ctx, user)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		return
	}
	fmt.Printf("Created user with ID: %d\n", user.ID)

	// Get user by ID
	retrievedUser, err := userRepo.GetByID(ctx, user.ID)
	if err != nil {
		log.Printf("Failed to get user by ID: %v", err)
		return
	}
	fmt.Printf("Retrieved user: %s (%s)\n", retrievedUser.Name, retrievedUser.Email)

	// Get user by email
	userByEmail, err := userRepo.GetByEmail(ctx, user.Email)
	if err != nil {
		log.Printf("Failed to get user by email: %v", err)
		return
	}
	fmt.Printf("User found by email: %s\n", userByEmail.Name)

	// Update user
	user.Name = "John Smith"
	err = userRepo.Update(ctx, user)
	if err != nil {
		log.Printf("Failed to update user: %v", err)
		return
	}
	fmt.Printf("Updated user name to: %s\n", user.Name)

	// Check if user exists
	exists, err := userRepo.Exists(ctx, user.ID)
	if err != nil {
		log.Printf("Failed to check if user exists: %v", err)
		return
	}
	fmt.Printf("User exists: %t\n", exists)

	// Check if user exists by email
	existsByEmail, err := userRepo.ExistsByEmail(ctx, user.Email)
	if err != nil {
		log.Printf("Failed to check if user exists by email: %v", err)
		return
	}
	fmt.Printf("User exists by email: %t\n", existsByEmail)

	// List users with pagination
	users, err := userRepo.List(ctx, 10, 0)
	if err != nil {
		log.Printf("Failed to list users: %v", err)
		return
	}
	fmt.Printf("Found %d users\n", len(users))

	// Count total users
	count, err := userRepo.Count(ctx)
	if err != nil {
		log.Printf("Failed to count users: %v", err)
		return
	}
	fmt.Printf("Total users: %d\n", count)

	// Delete user
	err = userRepo.Delete(ctx, user.ID)
	if err != nil {
		log.Printf("Failed to delete user: %v", err)
		return
	}
	fmt.Printf("Deleted user with ID: %d\n", user.ID)
}

// ExampleWithTransaction demonstrates how to use the repository with transactions
func ExampleWithTransaction() {
	// Create database configuration
	config := database.DefaultPostgreSQLConfig()

	// Create database connection
	db, err := database.NewDatabase(config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create user repository
	userRepo := NewUserRepository(db)

	ctx := context.Background()

	// Begin transaction
	tx, err := userRepo.BeginTx(ctx)
	if err != nil {
		log.Printf("Failed to begin transaction: %v", err)
		return
	}

	// Create multiple users in a transaction
	users := []*models.User{
		{Name: "Alice Johnson", Email: "alice@example.com"},
		{Name: "Bob Smith", Email: "bob@example.com"},
		{Name: "Charlie Brown", Email: "charlie@example.com"},
	}

	for _, user := range users {
		err = userRepo.Create(ctx, user)
		if err != nil {
			// Rollback transaction on error
			tx.Rollback()
			log.Printf("Failed to create user, rolling back: %v", err)
			return
		}
		fmt.Printf("Created user: %s\n", user.Name)
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return
	}

	fmt.Println("Transaction committed successfully")
}
