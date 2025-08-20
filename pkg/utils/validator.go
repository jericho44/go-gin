package utils

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator holds the validator instance and custom validation functions
type Validator struct {
	validate *validator.Validate
}

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// ValidationErrors represents a collection of validation errors
type ValidationErrors []ValidationError

// NewValidator creates a new validator instance with custom validation functions
func NewValidator() *Validator {
	validate := validator.New()

	// Register custom validation functions
	validate.RegisterValidation("username", validateUsername)
	validate.RegisterValidation("strong_password", validateStrongPassword)
	validate.RegisterValidation("phone", validatePhone)

	// Use JSON field names in validation errors
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{
		validate: validate,
	}
}

// ValidateStruct validates a struct and returns formatted validation errors
func (v *Validator) ValidateStruct(s interface{}) ValidationErrors {
	var validationErrors ValidationErrors

	err := v.validate.Struct(s)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			validationError := ValidationError{
				Field:   err.Field(),
				Tag:     err.Tag(),
				Value:   fmt.Sprintf("%v", err.Value()),
				Message: getErrorMessage(err),
			}
			validationErrors = append(validationErrors, validationError)
		}
	}

	return validationErrors
}

// ValidateVar validates a single variable
func (v *Validator) ValidateVar(field interface{}, tag string) error {
	return v.validate.Var(field, tag)
}

// HasErrors checks if there are any validation errors
func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}

// Error implements the error interface for ValidationErrors
func (ve ValidationErrors) Error() string {
	var messages []string
	for _, err := range ve {
		messages = append(messages, err.Message)
	}
	return strings.Join(messages, "; ")
}

// ToStringSlice converts validation errors to a slice of error messages
func (ve ValidationErrors) ToStringSlice() []string {
	var messages []string
	for _, err := range ve {
		messages = append(messages, err.Message)
	}
	return messages
}

// ToMap converts validation errors to a map with field names as keys
func (ve ValidationErrors) ToMap() map[string]string {
	errorMap := make(map[string]string)
	for _, err := range ve {
		errorMap[err.Field] = err.Message
	}
	return errorMap
}

// getErrorMessage returns a user-friendly error message for validation errors
func getErrorMessage(fe validator.FieldError) string {
	field := fe.Field()

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", field, fe.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", field, fe.Param())
	case "numeric":
		return fmt.Sprintf("%s must be a number", field)
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "username":
		return fmt.Sprintf("%s must be 3-30 characters long and contain only letters, numbers, underscores, and hyphens", field)
	case "strong_password":
		return fmt.Sprintf("%s must be at least 8 characters long and contain at least one uppercase letter, one lowercase letter, one number, and one special character", field)
	case "phone":
		return fmt.Sprintf("%s must be a valid phone number", field)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, fe.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, fe.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, fe.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, fe.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, fe.Param())
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// Custom validation functions

// validateUsername validates username format (3-30 chars, alphanumeric, underscore, hyphen)
func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	if len(username) < 3 || len(username) > 30 {
		return false
	}

	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, username)
	return matched
}

// validateStrongPassword validates password strength
func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	// Check for at least one uppercase letter
	hasUpper, _ := regexp.MatchString(`[A-Z]`, password)
	if !hasUpper {
		return false
	}

	// Check for at least one lowercase letter
	hasLower, _ := regexp.MatchString(`[a-z]`, password)
	if !hasLower {
		return false
	}

	// Check for at least one digit
	hasDigit, _ := regexp.MatchString(`\d`, password)
	if !hasDigit {
		return false
	}

	// Check for at least one special character
	hasSpecial, _ := regexp.MatchString(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`, password)
	if !hasSpecial {
		return false
	}

	return true
}

// validatePhone validates phone number format (basic validation)
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()

	// Remove common phone number characters
	cleanPhone := regexp.MustCompile(`[^\d]`).ReplaceAllString(phone, "")

	// Check if it's between 10-15 digits (international format)
	if len(cleanPhone) < 10 || len(cleanPhone) > 15 {
		return false
	}

	return true
}

// Global validator instance
var globalValidator *Validator

// init initializes the global validator
func init() {
	globalValidator = NewValidator()
}

// ValidateStruct validates a struct using the global validator instance
func ValidateStruct(s interface{}) ValidationErrors {
	return globalValidator.ValidateStruct(s)
}

// ValidateVar validates a single variable using the global validator instance
func ValidateVar(field interface{}, tag string) error {
	return globalValidator.ValidateVar(field, tag)
}
