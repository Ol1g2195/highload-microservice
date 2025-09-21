package handlers

import (
	"net/http"
	"strconv"

	"highload-microservice/internal/models"
	"highload-microservice/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type UserHandler struct {
	userService *services.UserService
	logger      *logrus.Logger
}

func NewUserHandler(userService *services.UserService, logger *logrus.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	h.logger.Info("CreateUser handler called")
	
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	h.logger.Infof("Creating user with email: %s", req.Email)
	user, err := h.userService.CreateUser(c.Request.Context(), req)
	if err != nil {
		h.logger.Errorf("Failed to create user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user", "details": err.Error()})
		return
	}

	h.logger.Infof("User created successfully: %s", user.ID)
	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Errorf("Invalid user ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), id)
	if err != nil {
		h.logger.Errorf("Failed to get user: %v", err)
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Errorf("Invalid user ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), id, req)
	if err != nil {
		h.logger.Errorf("Failed to update user: %v", err)
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Errorf("Invalid user ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.userService.DeleteUser(c.Request.Context(), id)
	if err != nil {
		h.logger.Errorf("Failed to delete user: %v", err)
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	users, err := h.userService.ListUsers(c.Request.Context(), page, limit)
	if err != nil {
		h.logger.Errorf("Failed to list users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

