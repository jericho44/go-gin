package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		config         CORSConfig
		origin         string
		method         string
		expectedOrigin string
		expectedStatus int
	}{
		{
			name: "Allow all origins with wildcard",
			config: CORSConfig{
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{"GET", "POST"},
				AllowedHeaders: []string{"Content-Type"},
			},
			origin:         "https://example.com",
			method:         "GET",
			expectedOrigin: "https://example.com",
			expectedStatus: http.StatusOK,
		},
		{
			name: "Allow specific origin",
			config: CORSConfig{
				AllowedOrigins: []string{"https://example.com"},
				AllowedMethods: []string{"GET", "POST"},
				AllowedHeaders: []string{"Content-Type"},
			},
			origin:         "https://example.com",
			method:         "GET",
			expectedOrigin: "https://example.com",
			expectedStatus: http.StatusOK,
		},
		{
			name: "Reject unauthorized origin",
			config: CORSConfig{
				AllowedOrigins: []string{"https://example.com"},
				AllowedMethods: []string{"GET", "POST"},
				AllowedHeaders: []string{"Content-Type"},
			},
			origin:         "https://malicious.com",
			method:         "GET",
			expectedOrigin: "",
			expectedStatus: http.StatusOK,
		},
		{
			name: "Handle OPTIONS preflight request",
			config: CORSConfig{
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{"GET", "POST"},
				AllowedHeaders: []string{"Content-Type"},
			},
			origin:         "https://example.com",
			method:         "OPTIONS",
			expectedOrigin: "https://example.com",
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(CORS(tt.config))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})
			router.POST("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req, _ := http.NewRequest(tt.method, "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedOrigin != "" {
				assert.Equal(t, tt.expectedOrigin, w.Header().Get("Access-Control-Allow-Origin"))
			} else {
				assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
			}

			// Check that CORS headers are set
			if tt.expectedOrigin != "" {
				assert.Equal(t, "GET, POST", w.Header().Get("Access-Control-Allow-Methods"))
				assert.Equal(t, "Content-Type", w.Header().Get("Access-Control-Allow-Headers"))
				assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
				assert.Equal(t, "86400", w.Header().Get("Access-Control-Max-Age"))
			}
		})
	}
}

func TestIsOriginAllowed(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		allowedOrigins []string
		expected       bool
	}{
		{
			name:           "Wildcard allows all",
			origin:         "https://example.com",
			allowedOrigins: []string{"*"},
			expected:       true,
		},
		{
			name:           "Exact match",
			origin:         "https://example.com",
			allowedOrigins: []string{"https://example.com", "https://other.com"},
			expected:       true,
		},
		{
			name:           "No match",
			origin:         "https://malicious.com",
			allowedOrigins: []string{"https://example.com", "https://other.com"},
			expected:       false,
		},
		{
			name:           "Wildcard subdomain match",
			origin:         "https://api.example.com",
			allowedOrigins: []string{"*.example.com"},
			expected:       true,
		},
		{
			name:           "Wildcard subdomain root domain match",
			origin:         "https://example.com",
			allowedOrigins: []string{"*.example.com"},
			expected:       true,
		},
		{
			name:           "Empty allowed origins",
			origin:         "https://example.com",
			allowedOrigins: []string{},
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isOriginAllowed(tt.origin, tt.allowedOrigins)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(DefaultCORS())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
}
