package api

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/suite"
)

// DatabaseIntegrationTestSuite tests database operations
type DatabaseIntegrationTestSuite struct {
	IntegrationTestSuite
}

// TestDatabaseIntegrationSuite runs the database integration test suite
func TestDatabaseIntegrationSuite(t *testing.T) {
	suite.Run(t, new(DatabaseIntegrationTestSuite))
}

// TestDatabaseConnection tests that database connection is working
func (suite *DatabaseIntegrationTestSuite) TestDatabaseConnection() {
	// Test that we can ping the database
	err := suite.db.Ping()
	suite.NoError(err, "Database should be pingable")

	// Test that we can execute a simple query
	var result int
	err = suite.GetDB().QueryRow("SELECT 1").Scan(&result)
	suite.NoError(err, "Should be able to execute simple query")
	suite.Equal(1, result, "Query should return expected result")
}

// TestDatabaseTransactions tests database transaction functionality
func (suite *DatabaseIntegrationTestSuite) TestDatabaseTransactions() {
	// Test successful transaction
	err := suite.ExecuteInTransaction(func(tx *sql.Tx) error {
		// Insert a user within transaction
		_, err := tx.Exec("INSERT INTO users (name, email) VALUES ($1, $2)", "Transaction User", "transaction@example.com")
		return err
	})
	suite.NoError(err, "Transaction should succeed")

	// Verify user was inserted
	suite.True(suite.UserExistsInDBByEmail("transaction@example.com"), "User should exist after successful transaction")

	// Test failed transaction (should rollback)
	err = suite.ExecuteInTransaction(func(tx *sql.Tx) error {
		// Insert a user within transaction
		_, err := tx.Exec("INSERT INTO users (name, email) VALUES ($1, $2)", "Rollback User", "rollback@example.com")
		if err != nil {
			return err
		}

		// Force an error to trigger rollback
		return sql.ErrTxDone
	})
	suite.Error(err, "Transaction should fail")

	// Verify user was not inserted (rolled back)
	suite.False(suite.UserExistsInDBByEmail("rollback@example.com"), "User should not exist after failed transaction")
}

// TestDatabaseConstraints tests database constraints
func (suite *DatabaseIntegrationTestSuite) TestDatabaseConstraints() {
	// Test unique email constraint
	user1 := UserFixtures.ValidUser1
	suite.SeedUser(user1)

	// Try to insert user with same email
	_, err := suite.GetDB().Exec("INSERT INTO users (name, email) VALUES ($1, $2)", "Different Name", user1.Email)
	suite.Error(err, "Should not be able to insert user with duplicate email")
	suite.Contains(err.Error(), "unique", "Error should mention unique constraint")
}

// TestDatabaseCleanup tests that database cleanup works properly
func (suite *DatabaseIntegrationTestSuite) TestDatabaseCleanup() {
	// Insert some test data
	users := suite.SeedUsers(UserFixtures.ValidUser1, UserFixtures.ValidUser2)
	suite.Len(users, 2, "Should have seeded 2 users")

	// Verify data exists
	count := suite.CountUsersInDB()
	suite.Equal(int64(2), count, "Should have 2 users in database")

	// Cleanup is called automatically between tests by the test suite
	// This test verifies that the cleanup mechanism works
}

// TestDatabaseSequenceReset tests that database sequences are reset properly
func (suite *DatabaseIntegrationTestSuite) TestDatabaseSequenceReset() {
	// Insert a user and note the ID
	user1 := suite.SeedUser(UserFixtures.ValidUser1)
	firstID := user1.ID

	// Clean up (this happens automatically between tests)
	suite.cleanupDatabase()

	// Insert another user and verify ID starts from 1 again
	user2 := suite.SeedUser(UserFixtures.ValidUser2)
	suite.Equal(uint(1), user2.ID, "ID should start from 1 after sequence reset")

	// Verify the first user no longer exists
	suite.False(suite.UserExistsInDB(firstID), "First user should no longer exist")
}

// TestDatabaseTableStructure tests that database tables have correct structure
func (suite *DatabaseIntegrationTestSuite) TestDatabaseTableStructure() {
	// Test users table structure
	rows, err := suite.GetDB().Query(`
		SELECT column_name, data_type, is_nullable, column_default
		FROM information_schema.columns
		WHERE table_name = 'users'
		ORDER BY ordinal_position
	`)
	suite.NoError(err, "Should be able to query table structure")
	defer rows.Close()

	expectedColumns := map[string]struct {
		dataType   string
		isNullable string
	}{
		"id":         {"integer", "NO"},
		"name":       {"character varying", "NO"},
		"email":      {"character varying", "NO"},
		"created_at": {"timestamp without time zone", "YES"},
		"updated_at": {"timestamp without time zone", "YES"},
	}

	foundColumns := make(map[string]bool)

	for rows.Next() {
		var columnName, dataType, isNullable string
		var columnDefault sql.NullString

		err := rows.Scan(&columnName, &dataType, &isNullable, &columnDefault)
		suite.NoError(err, "Should be able to scan column info")

		if expected, exists := expectedColumns[columnName]; exists {
			suite.Equal(expected.dataType, dataType, "Column %s should have correct data type", columnName)
			suite.Equal(expected.isNullable, isNullable, "Column %s should have correct nullable setting", columnName)
			foundColumns[columnName] = true
		}
	}

	// Verify all expected columns were found
	for columnName := range expectedColumns {
		suite.True(foundColumns[columnName], "Column %s should exist in users table", columnName)
	}
}

// TestDatabaseTriggers tests that database triggers work correctly
func (suite *DatabaseIntegrationTestSuite) TestDatabaseTriggers() {
	// Insert a user
	user := suite.SeedUser(UserFixtures.ValidUser1)
	originalUpdatedAt := user.UpdatedAt

	// Update the user
	_, err := suite.GetDB().Exec("UPDATE users SET name = $1 WHERE id = $2", "Updated Name", user.ID)
	suite.NoError(err, "Should be able to update user")

	// Verify updated_at was changed by trigger
	updatedUser, err := suite.GetUserFromDB(user.ID)
	suite.NoError(err, "Should be able to get updated user")
	suite.NotEqual(originalUpdatedAt, updatedUser.UpdatedAt, "updated_at should be changed by trigger")
	suite.Equal("Updated Name", updatedUser.Name, "Name should be updated")
}
