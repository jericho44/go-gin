package utils

import (
	"gin-golang-app/internal/models"

	"github.com/gin-gonic/gin"
)

// Example usage of the validation utilities

// ValidateUserInput demonstrates how to validate user input using the validation utilities
func ValidateUserInput(c *gin.Context, user *models.User) bool {
	// Validate the user struct using our custom validator
	validationErrors := ValidateStruct(user)

	if validationErrors.HasErrors() {
		// Use the ValidationErrorResponseFromValidator helper from response.go
		ValidationErrorResponseFromValidator(c, validationErrors)
		return false
	}

	return true
}

// ValidateUserInputWithCustomResponse demonstrates custom error response formatting
func ValidateUserInputWithCustomResponse(c *gin.Context, user *models.User) bool {
	// Validate the user struct
	validationErrors := ValidateStruct(user)

	if validationErrors.HasErrors() {
		// Create a custom response with detailed field errors
		response := APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   "Please check the following fields",
			Data:    validationErrors.ToMap(), // Returns map[string]string with field -> error message
		}
		c.JSON(400, response)
		return false
	}

	return true
}

// Example request structs with validation tags

// CreateUserRequestWithValidation demonstrates comprehensive validation
type CreateUserRequestWithValidation struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"username"`
	Password string `json:"password" validate:"strong_password"`
	Phone    string `json:"phone,omitempty" validate:"omitempty,phone"`
	Age      int    `json:"age,omitempty" validate:"omitempty,gte=0,lte=150"`
}

// ValidateCreateUserRequest demonstrates validation of request structs
func ValidateCreateUserRequest(c *gin.Context, req *CreateUserRequestWithValidation) bool {
	validationErrors := ValidateStruct(req)

	if validationErrors.HasErrors() {
		// Use the simple string slice version for basic validation errors
		ValidationErrorResponseSimple(c, validationErrors.ToStringSlice())
		return false
	}

	return true
}

// Example of validating individual fields

// ValidateEmailField demonstrates validating a single field
func ValidateEmailField(email string) error {
	return ValidateVar(email, "required,email")
}

// ValidateUsernameField demonstrates validating with custom validation
func ValidateUsernameField(username string) error {
	return ValidateVar(username, "required,username")
}

// ValidatePasswordField demonstrates password validation
func ValidatePasswordField(password string) error {
	return ValidateVar(password, "required,strong_password")
}

// Example handler method using the validation utilities
// This would typically be in a handler file, but shown here for demonstration

/*
func (h *UserHandler) CreateUserWithValidation(c *gin.Context) {
	var req CreateUserRequestWithValidation

	// Bind JSON to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequestResponse(c, "Invalid JSON", err.Error())
		return
	}

	// Validate using our custom validator
	if !ValidateCreateUserRequest(c, &req) {
		return // Response already sent by validation function
	}

	// Convert to domain model
	user := &models.User{
		Name:  req.Name,
		Email: req.Email,
	}

	// Additional validation on the domain model if needed
	if !ValidateUserInput(c, user) {
		return // Response already sent by validation function
	}

	// Proceed with business logic...
	createdUser, err := h.userService.CreateUser(c.Request.Context(), user)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	CreatedResponse(c, "User created successfully", createdUser)
}
*/
