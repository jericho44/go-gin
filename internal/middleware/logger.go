package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Logger creates a gin middleware for request logging with request ID tracking
func Logger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Get or generate request ID
		requestID := param.Keys["request_id"]
		if requestID == nil {
			requestID = uuid.New().String()
		}

		// Log the request details
		logrus.WithFields(logrus.Fields{
			"request_id":    requestID,
			"client_ip":     param.ClientIP,
			"method":        param.Method,
			"path":          param.Path,
			"status_code":   param.StatusCode,
			"latency":       param.Latency,
			"user_agent":    param.Request.UserAgent(),
			"response_size": param.BodySize,
		}).Info("HTTP Request")

		return ""
	})
}

// RequestID middleware adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID already exists in headers
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set request ID in context and response header
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// ResponseTimeLogger middleware logs response time for each request
func ResponseTimeLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logrus.WithFields(logrus.Fields{
			"method":      param.Method,
			"path":        param.Path,
			"status_code": param.StatusCode,
			"latency":     param.Latency,
			"client_ip":   param.ClientIP,
		}).Infof("Request completed in %v", param.Latency)

		return ""
	})
}

// CombinedLogger combines request ID generation, logging, and response time tracking
func CombinedLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Get or generate request ID
		requestID := param.Keys["request_id"]
		if requestID == nil {
			requestID = uuid.New().String()
		}

		// Log comprehensive request details
		logrus.WithFields(logrus.Fields{
			"request_id":    requestID,
			"client_ip":     param.ClientIP,
			"method":        param.Method,
			"path":          param.Path,
			"status_code":   param.StatusCode,
			"latency":       param.Latency,
			"user_agent":    param.Request.UserAgent(),
			"response_size": param.BodySize,
			"timestamp":     param.TimeStamp.Format("2006-01-02 15:04:05"),
		}).Info("HTTP Request Completed")

		return ""
	})
}
