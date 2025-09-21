package middleware

import (
	"strings"
	"time"

	"highload-microservice/internal/security"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// SecurityLoggingMiddleware provides security event logging
type SecurityLoggingMiddleware struct {
	auditor *security.SecurityAuditor
	logger  *logrus.Logger
}

// NewSecurityLoggingMiddleware creates a new security logging middleware
func NewSecurityLoggingMiddleware(auditor *security.SecurityAuditor, logger *logrus.Logger) *SecurityLoggingMiddleware {
	return &SecurityLoggingMiddleware{
		auditor: auditor,
		logger:  logger,
	}
}

// LogRequest logs all requests for security analysis
func (slm *SecurityLoggingMiddleware) LogRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Get request information
		ipAddress := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")
		requestID := c.GetString("request_id")
		endpoint := c.Request.URL.Path
		method := c.Request.Method

		// Process request
		c.Next()

		// Get response information
		status := c.Writer.Status()
		duration := time.Since(start)

		// Log request
		slm.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"method":     method,
			"endpoint":   endpoint,
			"ip_address": ipAddress,
			"user_agent": userAgent,
			"status":     status,
			"duration":   duration,
		}).Info("Request processed")

		// Log security events based on status
		if status >= 400 {
			slm.logSecurityEvent(c, status)
		}
	}
}

// LogAuthentication logs authentication events
func (slm *SecurityLoggingMiddleware) LogAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This will be called after authentication middleware
		// Log authentication success/failure
		userID, exists := c.Get("user_id")
		if exists {
			if userIDStr, ok := userID.(string); ok {
				if userUUID, err := uuid.Parse(userIDStr); err == nil {
					slm.auditor.LogLoginSuccess(
						userUUID,
						c.ClientIP(),
						c.GetHeader("User-Agent"),
						c.GetString("request_id"),
					)
				}
			}
		}

		c.Next()
	}
}

// LogAuthorization logs authorization events
func (slm *SecurityLoggingMiddleware) LogAuthorization() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This will be called after authorization middleware
		// Log access granted/denied
		userID, exists := c.Get("user_id")
		var userUUID *uuid.UUID
		if exists {
			if userIDStr, ok := userID.(string); ok {
				if parsed, err := uuid.Parse(userIDStr); err == nil {
					userUUID = &parsed
				}
			}
		}

		// Check if access was denied (status 403)
		if c.Writer.Status() == 403 {
			slm.auditor.LogAccessDenied(
				userUUID,
				c.ClientIP(),
				c.GetHeader("User-Agent"),
				c.GetString("request_id"),
				c.Request.URL.Path,
				"insufficient_permissions",
			)
		}

		c.Next()
	}
}

// LogRateLimit logs rate limiting events
func (slm *SecurityLoggingMiddleware) LogRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This will be called after rate limiting middleware
		// Check if request was rate limited (status 429)
		if c.Writer.Status() == 429 {
			slm.auditor.LogRateLimitExceeded(
				c.ClientIP(),
				c.GetHeader("User-Agent"),
				c.GetString("request_id"),
				c.Request.URL.Path,
				60, // Default limit
			)
		}

		c.Next()
	}
}

// LogDDoS logs DDoS protection events
func (slm *SecurityLoggingMiddleware) LogDDoS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This will be called after DDoS protection middleware
		// Check if request was blocked by DDoS protection
		if c.Writer.Status() == 429 && c.GetHeader("X-DDoS-Blocked") == "true" {
			slm.auditor.LogDDoSDetected(
				c.ClientIP(),
				c.GetHeader("User-Agent"),
				c.GetString("request_id"),
				100, // Default request count
			)
		}

		c.Next()
	}
}

// LogValidation logs validation events
func (slm *SecurityLoggingMiddleware) LogValidation() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This will be called after validation middleware
		// Check if validation failed (status 400)
		if c.Writer.Status() == 400 {
			// Get validation errors from context
			validationErrors, exists := c.Get("validation_errors")
			if exists {
				if errors, ok := validationErrors.([]string); ok {
					slm.auditor.LogValidationFailed(
						c.ClientIP(),
						c.GetHeader("User-Agent"),
						c.GetString("request_id"),
						c.Request.URL.Path,
						errors,
					)
				}
			}
		}

		c.Next()
	}
}

// LogSuspiciousInput logs suspicious input attempts
func (slm *SecurityLoggingMiddleware) LogSuspiciousInput() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for suspicious input patterns
		userAgent := c.GetHeader("User-Agent")
		if slm.isSuspiciousUserAgent(userAgent) {
			slm.auditor.LogEvent(security.SecurityEvent{
				EventType: security.EventTypeSuspiciousUserAgent,
				Severity:  security.SeverityMedium,
				IPAddress: c.ClientIP(),
				UserAgent: userAgent,
				RequestID: c.GetString("request_id"),
				Endpoint:  c.Request.URL.Path,
				Method:    c.Request.Method,
				Details: map[string]interface{}{
					"user_agent": userAgent,
					"reason":     "suspicious_user_agent",
				},
			})
		}

		c.Next()
	}
}

// LogAPIKeyUsage logs API key usage
func (slm *SecurityLoggingMiddleware) LogAPIKeyUsage() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This will be called after API key middleware
		// Log API key usage
		apiKeyID, exists := c.Get("api_key_id")
		if exists {
			if apiKeyIDStr, ok := apiKeyID.(string); ok {
				if apiKeyUUID, err := uuid.Parse(apiKeyIDStr); err == nil {
					userID, userExists := c.Get("user_id")
					var userUUID *uuid.UUID
					if userExists {
						if userIDStr, ok := userID.(string); ok {
							if parsed, err := uuid.Parse(userIDStr); err == nil {
								userUUID = &parsed
							}
						}
					}

					slm.auditor.LogAPIKeyUsage(
						apiKeyUUID,
						userUUID,
						c.ClientIP(),
						c.GetHeader("User-Agent"),
						c.GetString("request_id"),
						c.Request.URL.Path,
					)
				}
			}
		}

		c.Next()
	}
}

// logSecurityEvent logs a security event based on response status
func (slm *SecurityLoggingMiddleware) logSecurityEvent(c *gin.Context, status int) {
	eventType := security.EventTypeUnusualActivity
	severity := security.SeverityLow

	switch status {
	case 401:
		eventType = security.EventTypeAccessDenied
		severity = security.SeverityMedium
	case 403:
		eventType = security.EventTypeAccessDenied
		severity = security.SeverityMedium
	case 429:
		eventType = security.EventTypeRateLimitExceeded
		severity = security.SeverityHigh
	case 500:
		eventType = security.EventTypeUnusualActivity
		severity = security.SeverityHigh
	}

	slm.auditor.LogEvent(security.SecurityEvent{
		EventType: eventType,
		Severity:  severity,
		IPAddress: c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		RequestID: c.GetString("request_id"),
		Endpoint:  c.Request.URL.Path,
		Method:    c.Request.Method,
		Status:    status,
		Details: map[string]interface{}{
			"status_code": status,
			"reason":      "http_status_code",
		},
	})
}

// isSuspiciousUserAgent checks if user agent looks suspicious
func (slm *SecurityLoggingMiddleware) isSuspiciousUserAgent(userAgent string) bool {
	if userAgent == "" {
		return true
	}

	suspiciousPatterns := []string{
		"sqlmap", "nikto", "nmap", "masscan",
		"zap", "burp", "w3af", "havij",
		"acunetix", "nessus", "openvas",
		"metasploit", "curl/7.0", "wget/1.0",
	}

	userAgentLower := strings.ToLower(userAgent)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(userAgentLower, pattern) {
			return true
		}
	}

	return false
}
