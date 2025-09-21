package middleware

import (
	"net/http"
	"strings"

	"highload-microservice/internal/models"
	"highload-microservice/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type AuthMiddleware struct {
	authService *services.AuthService
	logger      *logrus.Logger
}

func NewAuthMiddleware(authService *services.AuthService, logger *logrus.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		logger:      logger,
	}
}

// RequireAuth middleware that requires JWT authentication
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)
		if token == "" {
			m.logger.Warn("Authentication failed: no token provided")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		claims, err := m.authService.ValidateToken(token)
		if err != nil {
			m.logger.Warnf("Authentication failed: invalid token - %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Add user info to context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("claims", claims)

		m.logger.Debugf("User authenticated: %s (%s)", claims.Email, claims.Role)
		c.Next()
	}
}

// RequireRole middleware that requires specific role
func (m *AuthMiddleware) RequireRole(requiredRole models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check if user is authenticated
		role, exists := c.Get("user_role")
		if !exists {
			m.logger.Warn("Authorization failed: user not authenticated")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		userRole, ok := role.(models.UserRole)
		if !ok {
			m.logger.Error("Authorization failed: invalid role type")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		// Check role permissions
		if !m.hasPermission(userRole, requiredRole) {
			m.logger.Warnf("Authorization failed: insufficient permissions - user: %s, required: %s", userRole, requiredRole)
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		m.logger.Debugf("User authorized: %s has permission for %s", userRole, requiredRole)
		c.Next()
	}
}

// RequireAPIKey middleware that requires API key authentication
func (m *AuthMiddleware) RequireAPIKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := m.extractAPIKey(c)
		if apiKey == "" {
			m.logger.Warn("API key authentication failed: no API key provided")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key required"})
			c.Abort()
			return
		}

		permissions, err := m.authService.ValidateAPIKey(c.Request.Context(), apiKey)
		if err != nil {
			m.logger.Warnf("API key authentication failed: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			c.Abort()
			return
		}

		// Add permissions to context
		c.Set("api_permissions", permissions)

		m.logger.Debugf("API key authenticated with permissions: %v", permissions)
		c.Next()
	}
}

// RequireAPIPermission middleware that requires specific API permission
func (m *AuthMiddleware) RequireAPIPermission(requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check if API key is authenticated
		permissions, exists := c.Get("api_permissions")
		if !exists {
			m.logger.Warn("API authorization failed: API key not authenticated")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key authentication required"})
			c.Abort()
			return
		}

		perms, ok := permissions.([]string)
		if !ok {
			m.logger.Error("API authorization failed: invalid permissions type")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			c.Abort()
			return
		}

		// Check if required permission exists
		hasPermission := false
		for _, perm := range perms {
			if perm == requiredPermission || perm == "*" { // * means all permissions
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			m.logger.Warnf("API authorization failed: missing permission %s, available: %v", requiredPermission, perms)
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient API permissions"})
			c.Abort()
			return
		}

		m.logger.Debugf("API key authorized for permission: %s", requiredPermission)
		c.Next()
	}
}

// OptionalAuth middleware that adds user info if token is provided but doesn't require it
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)
		if token == "" {
			c.Next()
			return
		}

		claims, err := m.authService.ValidateToken(token)
		if err != nil {
			m.logger.Debugf("Optional authentication failed: %v", err)
			c.Next()
			return
		}

		// Add user info to context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("claims", claims)

		m.logger.Debugf("Optional user authenticated: %s (%s)", claims.Email, claims.Role)
		c.Next()
	}
}

// Helper methods

func (m *AuthMiddleware) extractToken(c *gin.Context) string {
	// Try Authorization header first
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	// Try query parameter
	return c.Query("token")
}

func (m *AuthMiddleware) extractAPIKey(c *gin.Context) string {
	// Try X-API-Key header first
	apiKey := c.GetHeader("X-API-Key")
	if apiKey != "" {
		return apiKey
	}

	// Try Authorization header with API key format
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "apikey" {
			return parts[1]
		}
	}

	// Try query parameter
	return c.Query("api_key")
}

func (m *AuthMiddleware) hasPermission(userRole, requiredRole models.UserRole) bool {
	// Define role hierarchy
	roleHierarchy := map[models.UserRole]int{
		models.RoleReadOnly: 1,
		models.RoleUser:     2,
		models.RoleAdmin:    3,
	}

	userLevel, userExists := roleHierarchy[userRole]
	requiredLevel, requiredExists := roleHierarchy[requiredRole]

	if !userExists || !requiredExists {
		return false
	}

	// User must have equal or higher level than required
	return userLevel >= requiredLevel
}
