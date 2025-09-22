package validation

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

// CustomValidator wraps the validator with custom validation rules
type CustomValidator struct {
	validator *validator.Validate
}

// NewCustomValidator creates a new custom validator
func NewCustomValidator() (*CustomValidator, error) {
	v := validator.New()

	// Register custom validations
	if err := v.RegisterValidation("strong_password", validateStrongPassword); err != nil {
		return nil, fmt.Errorf("failed to register strong_password validation: %w", err)
	}
	if err := v.RegisterValidation("safe_string", validateSafeString); err != nil {
		return nil, fmt.Errorf("failed to register safe_string validation: %w", err)
	}
	if err := v.RegisterValidation("uuid", validateUUID); err != nil {
		return nil, fmt.Errorf("failed to register uuid validation: %w", err)
	}
	if err := v.RegisterValidation("email_domain", validateEmailDomain); err != nil {
		return nil, fmt.Errorf("failed to register email_domain validation: %w", err)
	}
	if err := v.RegisterValidation("no_sql_injection", validateNoSQLInjection); err != nil {
		return nil, fmt.Errorf("failed to register no_sql_injection validation: %w", err)
	}
	if err := v.RegisterValidation("no_xss", validateNoXSS); err != nil {
		return nil, fmt.Errorf("failed to register no_xss validation: %w", err)
	}

	return &CustomValidator{
		validator: v,
	}, nil
}

// Validate validates a struct
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

// ValidateVar validates a single variable
func (cv *CustomValidator) ValidateVar(field interface{}, tag string) error {
	return cv.validator.Var(field, tag)
}

// validateStrongPassword validates password strength
func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	// Minimum length
	if len(password) < 8 {
		return false
	}

	// Maximum length
	if len(password) > 128 {
		return false
	}

	// Check for at least one uppercase letter
	hasUpper := false
	// Check for at least one lowercase letter
	hasLower := false
	// Check for at least one digit
	hasDigit := false
	// Check for at least one special character
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	// Must have at least 3 of the 4 character types
	count := 0
	if hasUpper {
		count++
	}
	if hasLower {
		count++
	}
	if hasDigit {
		count++
	}
	if hasSpecial {
		count++
	}

	return count >= 3
}

// validateSafeString validates that string doesn't contain dangerous characters
func validateSafeString(fl validator.FieldLevel) bool {
	value := fl.Field().String()

	// Check for null bytes
	if strings.Contains(value, "\x00") {
		return false
	}

	// Check for control characters (except newline and tab)
	for _, char := range value {
		if char < 32 && char != '\n' && char != '\t' && char != '\r' {
			return false
		}
	}

	return true
}

// validateUUID validates UUID format
func validateUUID(fl validator.FieldLevel) bool {
	value := fl.Field().String()

	// UUID v4 pattern
	uuidPattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	return uuidPattern.MatchString(strings.ToLower(value))
}

// validateEmailDomain validates email domain
func validateEmailDomain(fl validator.FieldLevel) bool {
	email := fl.Field().String()

	// Basic email validation
	emailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailPattern.MatchString(email) {
		return false
	}

	// Extract domain
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	domain := parts[1]

	// Check for suspicious domains
	suspiciousDomains := []string{
		"tempmail.org",
		"10minutemail.com",
		"guerrillamail.com",
		"mailinator.com",
		"throwaway.email",
	}

	for _, suspicious := range suspiciousDomains {
		if strings.Contains(domain, suspicious) {
			return false
		}
	}

	return true
}

// validateNoSQLInjection validates that string doesn't contain SQL injection patterns
func validateNoSQLInjection(fl validator.FieldLevel) bool {
	value := strings.ToLower(fl.Field().String())

	// Common SQL injection patterns
	sqlPatterns := []string{
		"' or '1'='1",
		"' or 1=1--",
		"'; drop table",
		"union select",
		"insert into",
		"delete from",
		"update set",
		"drop table",
		"create table",
		"alter table",
		"exec(",
		"execute(",
		"script>",
		"<script",
		"javascript:",
		"vbscript:",
		"onload=",
		"onerror=",
		"onclick=",
	}

	for _, pattern := range sqlPatterns {
		if strings.Contains(value, pattern) {
			return false
		}
	}

	return true
}

// validateNoXSS validates that string doesn't contain XSS patterns
func validateNoXSS(fl validator.FieldLevel) bool {
	value := strings.ToLower(fl.Field().String())

	// Common XSS patterns
	xssPatterns := []string{
		"<script",
		"</script>",
		"javascript:",
		"vbscript:",
		"onload=",
		"onerror=",
		"onclick=",
		"onmouseover=",
		"onfocus=",
		"onblur=",
		"onchange=",
		"onsubmit=",
		"onreset=",
		"onkeydown=",
		"onkeyup=",
		"onkeypress=",
		"<iframe",
		"<object",
		"<embed",
		"<applet",
		"<meta",
		"<link",
		"<style",
		"expression(",
		"url(",
		"@import",
	}

	for _, pattern := range xssPatterns {
		if strings.Contains(value, pattern) {
			return false
		}
	}

	return true
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// GetValidationErrors returns formatted validation errors
func (cv *CustomValidator) GetValidationErrors(err error) []ValidationError {
	var errors []ValidationError

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, ValidationError{
				Field:   e.Field(),
				Tag:     e.Tag(),
				Value:   fmt.Sprintf("%v", e.Value()),
				Message: getErrorMessage(e),
			})
		}
	}

	return errors
}

// getErrorMessage returns a human-readable error message
func getErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return fmt.Sprintf("%s must be a valid email address", fe.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", fe.Field(), fe.Param())
	case "strong_password":
		return fmt.Sprintf("%s must be a strong password (8-128 chars, at least 3 of: uppercase, lowercase, digit, special)", fe.Field())
	case "safe_string":
		return fmt.Sprintf("%s contains unsafe characters", fe.Field())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", fe.Field())
	case "email_domain":
		return fmt.Sprintf("%s must be from a trusted domain", fe.Field())
	case "no_sql_injection":
		return fmt.Sprintf("%s contains potentially dangerous content", fe.Field())
	case "no_xss":
		return fmt.Sprintf("%s contains potentially dangerous content", fe.Field())
	default:
		return fmt.Sprintf("%s is invalid", fe.Field())
	}
}
