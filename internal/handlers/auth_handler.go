package handlers

import (
	"net/http"

	"highload-microservice/internal/models"
	"highload-microservice/internal/security"
	"highload-microservice/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type AuthHandler struct {
	authService     *services.AuthService
	securityAuditor *security.SecurityAuditor
	logger          *logrus.Logger
}

func NewAuthHandler(authService *services.AuthService, securityAuditor *security.SecurityAuditor, logger *logrus.Logger) *AuthHandler {
	return &AuthHandler{
		authService:     authService,
		securityAuditor: securityAuditor,
		logger:          logger,
	}
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Invalid login request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	response, err := h.authService.AuthenticateUser(c.Request.Context(), req)
	if err != nil {
		// Log failed login attempt
		h.securityAuditor.LogLoginFailure(
			req.Email,
			c.ClientIP(),
			c.GetHeader("User-Agent"),
			c.GetString("request_id"),
			err.Error(),
		)

		h.logger.Errorf("Login failed for email %s: %v", req.Email, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Log successful login
	h.securityAuditor.LogLoginSuccess(
		response.User.ID,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
		c.GetString("request_id"),
	)

	h.logger.Infof("User logged in successfully: %s", req.Email)
	c.JSON(http.StatusOK, response)
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Invalid refresh token request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	response, err := h.authService.RefreshToken(c.Request.Context(), req)
	if err != nil {
		h.logger.Errorf("Token refresh failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	h.logger.Info("Token refreshed successfully")
	c.JSON(http.StatusOK, response)
}

// CreateAPIKey handles API key creation
func (h *AuthHandler) CreateAPIKey(c *gin.Context) {
	var req models.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Invalid API key creation request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	response, err := h.authService.CreateAPIKey(c.Request.Context(), req)
	if err != nil {
		h.logger.Errorf("API key creation failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create API key"})
		return
	}

	h.logger.Infof("API key created successfully: %s", req.Name)
	c.JSON(http.StatusCreated, response)
}

// GetProfile returns current user profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		h.logger.Error("User ID not found in context")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	email, _ := c.Get("user_email")
	role, _ := c.Get("user_role")

	profile := gin.H{
		"user_id": userID,
		"email":   email,
		"role":    role,
	}

	c.JSON(http.StatusOK, profile)
}

// Logout handles user logout (in a stateless system, this is mainly for logging)
func (h *AuthHandler) Logout(c *gin.Context) {
	userEmail, exists := c.Get("user_email")
	if exists {
		h.logger.Infof("User logged out: %s", userEmail)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
