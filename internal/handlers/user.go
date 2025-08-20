package handlers

import (
	"errors"
	"strconv"
	"strings"

	"gin-golang-app/internal/models"
	"gin-golang-app/internal/services"
	"gin-golang-app/pkg/utils"

	"github.com/gin-gonic/gin"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService services.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

// CreateUser handles POST /users requests
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest

	// Bind and validate request body
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body", err.Error())
		return
	}

	// Create user model from request
	user := &models.User{
		Name:  req.Name,
		Email: req.Email,
	}

	// Call service to create user
	createdUser, err := h.userService.CreateUser(c.Request.Context(), user)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	utils.CreatedResponse(c, "User created successfully", createdUser)
}

// GetUser handles GET /users/:id requests
func (h *UserHandler) GetUser(c *gin.Context) {
	// Extract and validate user ID from URL parameter
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID", "User ID must be a valid number")
		return
	}

	// Call service to get user
	user, err := h.userService.GetUserByID(c.Request.Context(), uint(id))
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	utils.OKResponse(c, "User retrieved successfully", user)
}

// UpdateUser handles PUT /users/:id requests
func (h *UserHandler) UpdateUser(c *gin.Context) {
	// Extract and validate user ID from URL parameter
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID", "User ID must be a valid number")
		return
	}

	var req UpdateUserRequest

	// Bind and validate request body
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request body", err.Error())
		return
	}

	// Create user model from request
	user := &models.User{
		ID:    uint(id),
		Name:  req.Name,
		Email: req.Email,
	}

	// Call service to update user
	updatedUser, err := h.userService.UpdateUser(c.Request.Context(), user)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	utils.OKResponse(c, "User updated successfully", updatedUser)
}

// DeleteUser handles DELETE /users/:id requests
func (h *UserHandler) DeleteUser(c *gin.Context) {
	// Extract and validate user ID from URL parameter
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID", "User ID must be a valid number")
		return
	}

	// Call service to delete user
	err = h.userService.DeleteUser(c.Request.Context(), uint(id))
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	utils.NoContentResponse(c)
}

// ListUsers handles GET /users requests with pagination
func (h *UserHandler) ListUsers(c *gin.Context) {
	// Parse pagination parameters with defaults
	page := 1
	limit := 10

	if pageParam := c.Query("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
			page = p
		}
	}

	if limitParam := c.Query("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 {
			limit = l
		}
	}

	// Call service to list users
	users, total, err := h.userService.ListUsers(c.Request.Context(), page, limit)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	// Calculate pagination metadata
	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	pagination := utils.Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}

	utils.PaginatedSuccessResponse(c, "Users retrieved successfully", users, pagination)
}

// GetUserByEmail handles GET /users/email/:email requests
func (h *UserHandler) GetUserByEmail(c *gin.Context) {
	email := c.Param("email")
	if email == "" {
		utils.BadRequestResponse(c, "Invalid email", "Email parameter is required")
		return
	}

	// Call service to get user by email
	user, err := h.userService.GetUserByEmail(c.Request.Context(), email)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	utils.OKResponse(c, "User retrieved successfully", user)
}

// handleServiceError maps service errors to appropriate HTTP responses
func (h *UserHandler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrUserNotFound):
		utils.NotFoundResponse(c, "User not found")
	case errors.Is(err, services.ErrUserAlreadyExists):
		utils.BadRequestResponse(c, "User already exists", err.Error())
	case errors.Is(err, services.ErrInvalidInput):
		utils.BadRequestResponse(c, "Invalid input", err.Error())
	case errors.Is(err, services.ErrInvalidEmail):
		utils.BadRequestResponse(c, "Invalid email", err.Error())
	case errors.Is(err, services.ErrInvalidName):
		utils.BadRequestResponse(c, "Invalid name", err.Error())
	case errors.Is(err, services.ErrInvalidPagination):
		utils.BadRequestResponse(c, "Invalid pagination parameters", err.Error())
	default:
		// Log the error for debugging (in a real app, use proper logging)
		// For now, we'll include it in the response for development
		if strings.Contains(err.Error(), "validation") {
			utils.BadRequestResponse(c, "Validation error", err.Error())
		} else {
			utils.InternalServerErrorResponse(c, "Internal server error", err.Error())
		}
	}
}
