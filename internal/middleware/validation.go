package middleware

import (
	"net/http"
	"reflect"
	"strings"

	"highload-microservice/internal/validation"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ValidationMiddleware provides validation middleware
type ValidationMiddleware struct {
	validator *validation.CustomValidator
	logger    *logrus.Logger
}

// NewValidationMiddleware creates a new validation middleware
func NewValidationMiddleware(logger *logrus.Logger) *ValidationMiddleware {
	return &ValidationMiddleware{
		validator: func() *validation.CustomValidator {
			v, err := validation.NewCustomValidator()
			if err != nil {
				logger.Fatalf("Failed to create custom validator: %v", err)
			}
			return v
		}(),
		logger: logger,
	}
}

// ValidateStruct validates a struct and returns errors
func (vm *ValidationMiddleware) ValidateStruct(obj interface{}) []validation.ValidationError {
	if err := vm.validator.Validate(obj); err != nil {
		return vm.validator.GetValidationErrors(err)
	}
	return nil
}

// ValidateRequest validates request body and returns validation errors
func (vm *ValidationMiddleware) ValidateRequest(obj interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a fresh instance per request based on provided type
		t := reflect.TypeOf(obj)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		newVal := reflect.New(t).Interface()

		// Bind JSON to new instance
		if err := c.ShouldBindJSON(newVal); err != nil {
			vm.logger.Warnf("Request binding failed: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid request format",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Validate struct
		if errors := vm.ValidateStruct(newVal); len(errors) > 0 {
			vm.logger.Warnf("Validation failed for %s: %v", c.Request.URL.Path, errors)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation failed",
				"details": errors,
			})
			c.Abort()
			return
		}

		// Store validated object in context
		c.Set("validated_data", newVal)
		c.Next()
	}
}

// ValidateQuery validates query parameters
func (vm *ValidationMiddleware) ValidateQuery(obj interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Bind query parameters to struct
		if err := c.ShouldBindQuery(obj); err != nil {
			vm.logger.Warnf("Query binding failed: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid query parameters",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Validate struct
		if errors := vm.ValidateStruct(obj); len(errors) > 0 {
			vm.logger.Warnf("Query validation failed for %s: %v", c.Request.URL.Path, errors)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid query parameters",
				"details": errors,
			})
			c.Abort()
			return
		}

		// Store validated object in context
		c.Set("validated_query", obj)
		c.Next()
	}
}

// SanitizeInput sanitizes input to prevent injection attacks
func (vm *ValidationMiddleware) SanitizeInput() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Sanitize query parameters
		for key, values := range c.Request.URL.Query() {
			for i, value := range values {
				values[i] = vm.sanitizeString(value)
			}
			c.Request.URL.RawQuery = strings.ReplaceAll(c.Request.URL.RawQuery, key+"="+values[0], key+"="+values[0])
		}

		// Sanitize path parameters
		for _, param := range c.Params {
			param.Value = vm.sanitizeString(param.Value)
		}

		c.Next()
	}
}

// sanitizeString removes potentially dangerous characters
func (vm *ValidationMiddleware) sanitizeString(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove control characters (except newline and tab)
	var result strings.Builder
	for _, char := range input {
		if char >= 32 || char == '\n' || char == '\t' || char == '\r' {
			result.WriteRune(char)
		}
	}

	return result.String()
}

// ValidateFileUpload validates file uploads
func (vm *ValidationMiddleware) ValidateFileUpload(maxSize int64, allowedTypes []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			vm.logger.Warnf("File upload failed: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "No file uploaded",
			})
			c.Abort()
			return
		}
		defer func() { _ = file.Close() }()

		// Check file size
		if header.Size > maxSize {
			vm.logger.Warnf("File too large: %d bytes (max: %d)", header.Size, maxSize)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":    "File too large",
				"max_size": maxSize,
			})
			c.Abort()
			return
		}

		// Check file type
		contentType := header.Header.Get("Content-Type")
		allowed := false
		for _, allowedType := range allowedTypes {
			if contentType == allowedType {
				allowed = true
				break
			}
		}

		if !allowed {
			vm.logger.Warnf("Invalid file type: %s", contentType)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":         "Invalid file type",
				"allowed_types": allowedTypes,
			})
			c.Abort()
			return
		}

		// Store file info in context
		c.Set("uploaded_file", file)
		c.Set("file_header", header)
		c.Next()
	}
}

// ValidatePagination validates pagination parameters
func (vm *ValidationMiddleware) ValidatePagination() gin.HandlerFunc {
	return func(c *gin.Context) {
		page := c.DefaultQuery("page", "1")
		limit := c.DefaultQuery("limit", "10")

		// Validate page
		if page == "0" || page == "" {
			page = "1"
		}

		// Validate limit
		if limit == "0" || limit == "" {
			limit = "10"
		}

		// Set sanitized values
		c.Request.URL.RawQuery = strings.ReplaceAll(c.Request.URL.RawQuery, "page="+page, "page="+page)
		c.Request.URL.RawQuery = strings.ReplaceAll(c.Request.URL.RawQuery, "limit="+limit, "limit="+limit)

		c.Next()
	}
}
