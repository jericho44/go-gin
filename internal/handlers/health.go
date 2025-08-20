package handlers

import (
	"net/http"
	"runtime"
	"time"

	"gin-golang-app/internal/database"
	"gin-golang-app/pkg/utils"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db *database.Database
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *database.Database) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

// HealthStatus represents the health status response
type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version"`
	Uptime    string                 `json:"uptime"`
	System    SystemInfo             `json:"system"`
	Database  DatabaseHealthStatus   `json:"database"`
	Services  map[string]interface{} `json:"services"`
}

// SystemInfo contains system information
type SystemInfo struct {
	GoVersion    string `json:"go_version"`
	NumGoroutine int    `json:"num_goroutine"`
	NumCPU       int    `json:"num_cpu"`
}

// DatabaseHealthStatus contains database health information
type DatabaseHealthStatus struct {
	Status         string `json:"status"`
	Driver         string `json:"driver"`
	OpenConns      int    `json:"open_connections"`
	InUseConns     int    `json:"in_use_connections"`
	IdleConns      int    `json:"idle_connections"`
	ResponseTimeMs int64  `json:"response_time_ms"`
	Error          string `json:"error,omitempty"`
}

var startTime = time.Now()

// HealthCheck handles GET /health requests
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	status := "healthy"
	httpStatus := http.StatusOK

	// Check database health
	dbHealth := h.checkDatabaseHealth()
	if dbHealth.Status != "healthy" {
		status = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	}

	// Prepare health status response
	healthStatus := HealthStatus{
		Status:    status,
		Timestamp: time.Now(),
		Version:   "1.0.0", // This could be injected from build info
		Uptime:    time.Since(startTime).String(),
		System: SystemInfo{
			GoVersion:    runtime.Version(),
			NumGoroutine: runtime.NumGoroutine(),
			NumCPU:       runtime.NumCPU(),
		},
		Database: dbHealth,
		Services: map[string]interface{}{
			"api": map[string]string{
				"status": "healthy",
			},
		},
	}

	if status == "healthy" {
		utils.SuccessResponse(c, httpStatus, "System is healthy", healthStatus)
	} else {
		utils.ErrorResponseWithCode(c, httpStatus, "System is unhealthy", "One or more services are not responding")
	}
}

// checkDatabaseHealth performs database health check
func (h *HealthHandler) checkDatabaseHealth() DatabaseHealthStatus {
	if h.db == nil {
		return DatabaseHealthStatus{
			Status: "unhealthy",
			Error:  "Database connection not initialized",
		}
	}

	if h.db.DB == nil {
		return DatabaseHealthStatus{
			Status: "unhealthy",
			Driver: h.db.Config.Driver,
			Error:  "Database connection is nil",
		}
	}

	start := time.Now()
	err := h.db.HealthCheck()
	responseTime := time.Since(start).Milliseconds()

	stats := h.db.GetStats()

	dbHealth := DatabaseHealthStatus{
		Driver:         h.db.Config.Driver,
		OpenConns:      stats.OpenConnections,
		InUseConns:     stats.InUse,
		IdleConns:      stats.Idle,
		ResponseTimeMs: responseTime,
	}

	if err != nil {
		dbHealth.Status = "unhealthy"
		dbHealth.Error = err.Error()
	} else {
		dbHealth.Status = "healthy"
	}

	return dbHealth
}

// ReadinessCheck handles GET /ready requests
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	// Check if all critical services are ready
	ready := true
	services := make(map[string]interface{})

	// Check database readiness
	dbHealth := h.checkDatabaseHealth()
	services["database"] = dbHealth

	if dbHealth.Status != "healthy" {
		ready = false
	}

	status := "ready"
	httpStatus := http.StatusOK

	if !ready {
		status = "not ready"
		httpStatus = http.StatusServiceUnavailable
	}

	response := map[string]interface{}{
		"status":    status,
		"timestamp": time.Now(),
		"services":  services,
	}

	if ready {
		utils.SuccessResponse(c, httpStatus, "Service is ready", response)
	} else {
		utils.ErrorResponseWithCode(c, httpStatus, "Service is not ready", "One or more dependencies are not ready")
	}
}

// LivenessCheck handles GET /live requests
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	// Simple liveness check - if we can respond, we're alive
	response := map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now(),
		"uptime":    time.Since(startTime).String(),
	}

	utils.OKResponse(c, "Service is alive", response)
}
