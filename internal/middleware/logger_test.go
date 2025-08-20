package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Capture log output
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	logrus.SetLevel(logrus.InfoLevel)

	router := gin.New()
	router.Use(RequestID()) // Add RequestID middleware first
	router.Use(Logger())    // Then add Logger middleware
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check that log was written
	logOutput := buf.String()
	assert.Contains(t, logOutput, "HTTP Request")
	assert.Contains(t, logOutput, "GET")
	assert.Contains(t, logOutput, "/test")
	assert.Contains(t, logOutput, "200")
	assert.Contains(t, logOutput, "test-agent")
}

func TestRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		existingRequestID  string
		expectNewRequestID bool
	}{
		{
			name:               "Generate new request ID when none exists",
			existingRequestID:  "",
			expectNewRequestID: true,
		},
		{
			name:               "Use existing request ID from header",
			existingRequestID:  "existing-request-id",
			expectNewRequestID: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(RequestID())
			router.GET("/test", func(c *gin.Context) {
				requestID, exists := c.Get("request_id")
				assert.True(t, exists)
				assert.NotEmpty(t, requestID)

				if tt.expectNewRequestID {
					assert.NotEqual(t, tt.existingRequestID, requestID)
				} else {
					assert.Equal(t, tt.existingRequestID, requestID)
				}

				c.JSON(http.StatusOK, gin.H{"request_id": requestID})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.existingRequestID != "" {
				req.Header.Set("X-Request-ID", tt.existingRequestID)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			// Check that X-Request-ID header is set in response
			responseRequestID := w.Header().Get("X-Request-ID")
			assert.NotEmpty(t, responseRequestID)

			if !tt.expectNewRequestID {
				assert.Equal(t, tt.existingRequestID, responseRequestID)
			}
		})
	}
}

func TestResponseTimeLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Capture log output
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	logrus.SetLevel(logrus.InfoLevel)

	router := gin.New()
	router.Use(ResponseTimeLogger())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check that log contains response time information
	logOutput := buf.String()
	assert.Contains(t, logOutput, "Request completed in")
	assert.Contains(t, logOutput, "GET")
	assert.Contains(t, logOutput, "/test")
	assert.Contains(t, logOutput, "200")
}

func TestCombinedLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Capture log output
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	logrus.SetLevel(logrus.InfoLevel)

	router := gin.New()
	router.Use(RequestID())      // Add RequestID middleware first
	router.Use(CombinedLogger()) // Then add CombinedLogger middleware
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "combined-test-agent")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check that comprehensive log was written
	logOutput := buf.String()
	assert.Contains(t, logOutput, "HTTP Request Completed")
	assert.Contains(t, logOutput, "GET")
	assert.Contains(t, logOutput, "/test")
	assert.Contains(t, logOutput, "200")
	assert.Contains(t, logOutput, "combined-test-agent")
	assert.Contains(t, logOutput, "request_id")
	assert.Contains(t, logOutput, "latency")
	assert.Contains(t, logOutput, "timestamp")
}

func TestLoggerWithDifferentStatusCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Capture log output
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	logrus.SetLevel(logrus.InfoLevel)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{
			name:           "Success request",
			path:           "/success",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Not found request",
			path:           "/notfound",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset() // Clear buffer for each test

			router := gin.New()
			router.Use(RequestID())
			router.Use(Logger())
			router.GET("/success", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check that the correct status code is logged
			logOutput := buf.String()
			if tt.expectedStatus == http.StatusOK {
				assert.Contains(t, logOutput, "200")
			} else {
				assert.Contains(t, logOutput, "404")
			}
		})
	}
}
