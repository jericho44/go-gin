package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestUser struct for testing validation
type TestUser struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"username"`
	Password string `json:"password" validate:"strong_password"`
	Phone    string `json:"phone" validate:"phone"`
	Age      int    `json:"age" validate:"gte=0,lte=150"`
}

func TestNewValidator(t *testing.T) {
	validator := NewValidator()
	assert.NotNil(t, validator)
	assert.NotNil(t, validator.validate)
}

func TestValidateStruct_ValidData(t *testing.T) {
	user := TestUser{
		Name:     "John Doe",
		Email:    "john@example.com",
		Username: "john_doe",
		Password: "StrongPass123!",
		Phone:    "+1234567890",
		Age:      25,
	}

	validator := NewValidator()
	errors := validator.ValidateStruct(user)

	assert.False(t, errors.HasErrors())
	assert.Equal(t, 0, len(errors))
}

func TestValidateStruct_InvalidData(t *testing.T) {
	user := TestUser{
		Name:     "",              // required field missing
		Email:    "invalid-email", // invalid email format
		Username: "ab",            // too short
		Password: "weak",          // weak password
		Phone:    "123",           // invalid phone
		Age:      -5,              // negative age
	}

	validator := NewValidator()
	errors := validator.ValidateStruct(user)

	assert.True(t, errors.HasErrors())
	assert.Equal(t, 6, len(errors))

	// Check specific error messages
	errorMap := errors.ToMap()
	assert.Contains(t, errorMap["name"], "required")
	assert.Contains(t, errorMap["email"], "valid email")
	assert.Contains(t, errorMap["username"], "3-30 characters")
	assert.Contains(t, errorMap["password"], "8 characters")
	assert.Contains(t, errorMap["phone"], "valid phone")
	assert.Contains(t, errorMap["age"], "greater than or equal")
}

func TestValidateStruct_RequiredFields(t *testing.T) {
	user := TestUser{}

	validator := NewValidator()
	errors := validator.ValidateStruct(user)

	assert.True(t, errors.HasErrors())

	// Should have errors for required fields
	errorMap := errors.ToMap()
	assert.Contains(t, errorMap, "name")
	assert.Contains(t, errorMap, "email")
}

func TestValidateVar(t *testing.T) {
	validator := NewValidator()

	// Test valid email
	err := validator.ValidateVar("test@example.com", "email")
	assert.NoError(t, err)

	// Test invalid email
	err = validator.ValidateVar("invalid-email", "email")
	assert.Error(t, err)

	// Test required field
	err = validator.ValidateVar("", "required")
	assert.Error(t, err)

	err = validator.ValidateVar("value", "required")
	assert.NoError(t, err)
}

func TestValidationErrors_HasErrors(t *testing.T) {
	var errors ValidationErrors

	assert.False(t, errors.HasErrors())

	errors = append(errors, ValidationError{
		Field:   "name",
		Message: "Name is required",
	})

	assert.True(t, errors.HasErrors())
}

func TestValidationErrors_Error(t *testing.T) {
	errors := ValidationErrors{
		{Field: "name", Message: "Name is required"},
		{Field: "email", Message: "Email is invalid"},
	}

	errorString := errors.Error()
	assert.Contains(t, errorString, "Name is required")
	assert.Contains(t, errorString, "Email is invalid")
	assert.Contains(t, errorString, ";")
}

func TestValidationErrors_ToStringSlice(t *testing.T) {
	errors := ValidationErrors{
		{Field: "name", Message: "Name is required"},
		{Field: "email", Message: "Email is invalid"},
	}

	messages := errors.ToStringSlice()
	assert.Equal(t, 2, len(messages))
	assert.Contains(t, messages, "Name is required")
	assert.Contains(t, messages, "Email is invalid")
}

func TestValidationErrors_ToMap(t *testing.T) {
	errors := ValidationErrors{
		{Field: "name", Message: "Name is required"},
		{Field: "email", Message: "Email is invalid"},
	}

	errorMap := errors.ToMap()
	assert.Equal(t, 2, len(errorMap))
	assert.Equal(t, "Name is required", errorMap["name"])
	assert.Equal(t, "Email is invalid", errorMap["email"])
}

func TestCustomValidation_Username(t *testing.T) {
	validator := NewValidator()

	// Valid usernames
	validUsernames := []string{"john_doe", "user123", "test-user", "abc"}
	for _, username := range validUsernames {
		err := validator.ValidateVar(username, "username")
		assert.NoError(t, err, "Username %s should be valid", username)
	}

	// Invalid usernames
	invalidUsernames := []string{"ab", "a", "", "user@name", "user name", "this_is_a_very_long_username_that_exceeds_limit"}
	for _, username := range invalidUsernames {
		err := validator.ValidateVar(username, "username")
		assert.Error(t, err, "Username %s should be invalid", username)
	}
}

func TestCustomValidation_StrongPassword(t *testing.T) {
	validator := NewValidator()

	// Valid passwords
	validPasswords := []string{"StrongPass123!", "MyP@ssw0rd", "Secure123#", "Complex1$"}
	for _, password := range validPasswords {
		err := validator.ValidateVar(password, "strong_password")
		assert.NoError(t, err, "Password %s should be valid", password)
	}

	// Invalid passwords
	invalidPasswords := []string{
		"weak",         // too short
		"password",     // no uppercase, no numbers, no special chars
		"PASSWORD",     // no lowercase, no numbers, no special chars
		"Password",     // no numbers, no special chars
		"Password123",  // no special chars
		"Password!",    // no numbers
		"password123!", // no uppercase
	}
	for _, password := range invalidPasswords {
		err := validator.ValidateVar(password, "strong_password")
		assert.Error(t, err, "Password %s should be invalid", password)
	}
}

func TestCustomValidation_Phone(t *testing.T) {
	validator := NewValidator()

	// Valid phone numbers
	validPhones := []string{
		"+1234567890",
		"1234567890",
		"+1-234-567-8900",
		"(123) 456-7890",
		"+44 20 7946 0958",
		"123-456-7890",
	}
	for _, phone := range validPhones {
		err := validator.ValidateVar(phone, "phone")
		assert.NoError(t, err, "Phone %s should be valid", phone)
	}

	// Invalid phone numbers
	invalidPhones := []string{
		"123",              // too short
		"12345",            // too short
		"",                 // empty
		"abcdefghij",       // no digits
		"1234567890123456", // too long
	}
	for _, phone := range invalidPhones {
		err := validator.ValidateVar(phone, "phone")
		assert.Error(t, err, "Phone %s should be invalid", phone)
	}
}

func TestGlobalValidator(t *testing.T) {
	user := TestUser{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	// Test global ValidateStruct function
	errors := ValidateStruct(user)
	assert.True(t, errors.HasErrors()) // Should have errors for missing required fields

	// Test global ValidateVar function
	err := ValidateVar("test@example.com", "email")
	assert.NoError(t, err)

	err = ValidateVar("invalid-email", "email")
	assert.Error(t, err)
}

func TestGetErrorMessage(t *testing.T) {
	validator := NewValidator()

	// Test with a struct that will generate various validation errors
	testStruct := struct {
		Required string `validate:"required"`
		Email    string `validate:"email"`
		Min      string `validate:"min=5"`
		Max      string `validate:"max=10"`
		Numeric  string `validate:"numeric"`
		Alpha    string `validate:"alpha"`
		Username string `validate:"username"`
		Password string `validate:"strong_password"`
		Phone    string `validate:"phone"`
	}{
		Required: "",
		Email:    "invalid",
		Min:      "abc",
		Max:      "this is too long",
		Numeric:  "abc",
		Alpha:    "123",
		Username: "ab",
		Password: "weak",
		Phone:    "123",
	}

	errors := validator.ValidateStruct(testStruct)
	assert.True(t, errors.HasErrors())

	// Check that error messages are user-friendly
	for _, err := range errors {
		assert.NotEmpty(t, err.Message)
		assert.NotContains(t, err.Message, "Key:")
		assert.NotContains(t, err.Message, "Error:")
	}
}
