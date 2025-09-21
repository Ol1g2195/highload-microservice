package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// SecurityConfig holds configuration for security middleware
type SecurityConfig struct {
	AllowedOrigins        []string
	AllowedMethods        []string
	AllowedHeaders        []string
	ExposedHeaders        []string
	AllowCredentials      bool
	MaxAge                int
	ContentTypeNosniff    bool
	FrameDeny             bool
	XSSProtection         bool
	ReferrerPolicy        string
	PermissionsPolicy     string
	ContentSecurityPolicy string
}

// SecurityMiddleware provides security headers and CORS
type SecurityMiddleware struct {
	config SecurityConfig
	logger *logrus.Logger
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(config SecurityConfig, logger *logrus.Logger) *SecurityMiddleware {
	return &SecurityMiddleware{
		config: config,
		logger: logger,
	}
}

// SecurityHeaders adds security headers to responses
func (sm *SecurityMiddleware) SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Content Security Policy
		if sm.config.ContentSecurityPolicy != "" {
			c.Header("Content-Security-Policy", sm.config.ContentSecurityPolicy)
		}

		// X-Content-Type-Options
		if sm.config.ContentTypeNosniff {
			c.Header("X-Content-Type-Options", "nosniff")
		}

		// X-Frame-Options
		if sm.config.FrameDeny {
			c.Header("X-Frame-Options", "DENY")
		}

		// X-XSS-Protection
		if sm.config.XSSProtection {
			c.Header("X-XSS-Protection", "1; mode=block")
		}

		// Referrer-Policy
		if sm.config.ReferrerPolicy != "" {
			c.Header("Referrer-Policy", sm.config.ReferrerPolicy)
		}

		// Permissions-Policy
		if sm.config.PermissionsPolicy != "" {
			c.Header("Permissions-Policy", sm.config.PermissionsPolicy)
		}

		// Strict-Transport-Security (HSTS)
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// X-Permitted-Cross-Domain-Policies
		c.Header("X-Permitted-Cross-Domain-Policies", "none")

		// Cross-Origin-Embedder-Policy
		c.Header("Cross-Origin-Embedder-Policy", "require-corp")

		// Cross-Origin-Opener-Policy
		c.Header("Cross-Origin-Opener-Policy", "same-origin")

		// Cross-Origin-Resource-Policy
		c.Header("Cross-Origin-Resource-Policy", "same-origin")

		// Cache-Control for sensitive endpoints
		if strings.HasPrefix(c.Request.URL.Path, "/api/v1/auth") {
			c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}

		// Remove server information
		c.Header("Server", "")

		// Remove X-Powered-By
		c.Header("X-Powered-By", "")

		c.Next()
	}
}

// CORS handles Cross-Origin Resource Sharing
func (sm *SecurityMiddleware) CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		allowed := false
		if len(sm.config.AllowedOrigins) == 0 {
			allowed = true // Allow all origins if none specified
		} else {
			for _, allowedOrigin := range sm.config.AllowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.Header("Access-Control-Allow-Methods", strings.Join(sm.config.AllowedMethods, ", "))
			c.Header("Access-Control-Allow-Headers", strings.Join(sm.config.AllowedHeaders, ", "))
			c.Header("Access-Control-Expose-Headers", strings.Join(sm.config.ExposedHeaders, ", "))
			c.Header("Access-Control-Max-Age", string(rune(sm.config.MaxAge)))

			if sm.config.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}

			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		// Add CORS headers for actual requests
		if allowed {
			c.Header("Access-Control-Allow-Methods", strings.Join(sm.config.AllowedMethods, ", "))
			c.Header("Access-Control-Allow-Headers", strings.Join(sm.config.AllowedHeaders, ", "))
			c.Header("Access-Control-Expose-Headers", strings.Join(sm.config.ExposedHeaders, ", "))

			if sm.config.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
		}

		c.Next()
	}
}

// RequestID adds a unique request ID to each request
func (sm *SecurityMiddleware) RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)

		c.Next()
	}
}

// SecurityLogging logs security-related events
func (sm *SecurityMiddleware) SecurityLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log suspicious requests
		userAgent := c.GetHeader("User-Agent")
		if isSuspiciousUserAgent(userAgent) {
			sm.logger.Warnf("Suspicious User-Agent detected: %s from IP: %s", userAgent, c.ClientIP())
		}

		// Log requests to sensitive endpoints
		if strings.HasPrefix(c.Request.URL.Path, "/api/v1/auth") {
			sm.logger.Infof("Authentication request from IP: %s to %s", c.ClientIP(), c.Request.URL.Path)
		}

		// Log admin requests
		if strings.HasPrefix(c.Request.URL.Path, "/admin") {
			sm.logger.Infof("Admin request from IP: %s to %s", c.ClientIP(), c.Request.URL.Path)
		}

		c.Next()
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	// Simple implementation - in production, use a proper UUID generator
	return "req_" + randomString(16)
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}

// isSuspiciousUserAgent checks if user agent looks suspicious
func isSuspiciousUserAgent(userAgent string) bool {
	if userAgent == "" {
		return true
	}

	suspiciousPatterns := []string{
		"sqlmap",
		"nikto",
		"nmap",
		"masscan",
		"zap",
		"burp",
		"w3af",
		"havij",
		"acunetix",
		"nessus",
		"openvas",
		"metasploit",
		"curl/7.0", // Very old curl version
		"wget/1.0", // Very old wget version
	}

	userAgentLower := strings.ToLower(userAgent)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(userAgentLower, pattern) {
			return true
		}
	}

	return false
}

// DefaultSecurityConfig returns a secure default configuration
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		AllowedOrigins: []string{
			"https://localhost:3000",
			"https://127.0.0.1:3000",
			"https://localhost:8080",
			"https://127.0.0.1:8080",
		},
		AllowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
			"HEAD",
		},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"X-Request-ID",
			"X-API-Key",
		},
		ExposedHeaders: []string{
			"X-Request-ID",
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
		},
		AllowCredentials:      true,
		MaxAge:                86400, // 24 hours
		ContentTypeNosniff:    true,
		FrameDeny:             true,
		XSSProtection:         true,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		PermissionsPolicy:     "geolocation=(), microphone=(), camera=()",
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self';",
	}
}
