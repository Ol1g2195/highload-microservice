package security

import (
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// SecurityEvent represents a security-related event
type SecurityEvent struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	EventType SecurityEventType      `json:"event_type"`
	Severity  SecuritySeverity       `json:"severity"`
	UserID    *uuid.UUID             `json:"user_id,omitempty"`
	IPAddress string                 `json:"ip_address"`
	UserAgent string                 `json:"user_agent"`
	RequestID string                 `json:"request_id"`
	Endpoint  string                 `json:"endpoint"`
	Method    string                 `json:"method"`
	Status    int                    `json:"status"`
	Details   map[string]interface{} `json:"details"`
	RiskScore int                    `json:"risk_score"`
	Blocked   bool                   `json:"blocked"`
}

// SecurityEventType represents the type of security event
type SecurityEventType string

const (
	// Authentication events
	EventTypeLoginSuccess SecurityEventType = "login_success"
	EventTypeLoginFailure SecurityEventType = "login_failure"
	EventTypeLogout       SecurityEventType = "logout"
	EventTypeTokenRefresh SecurityEventType = "token_refresh"
	EventTypeTokenExpired SecurityEventType = "token_expired"
	EventTypeInvalidToken SecurityEventType = "invalid_token"

	// Authorization events
	EventTypeAccessGranted       SecurityEventType = "access_granted"
	EventTypeAccessDenied        SecurityEventType = "access_denied"
	EventTypePrivilegeEscalation SecurityEventType = "privilege_escalation"

	// Rate limiting events
	EventTypeRateLimitExceeded SecurityEventType = "rate_limit_exceeded"
	EventTypeDDoSDetected      SecurityEventType = "ddos_detected"
	EventTypeIPBlocked         SecurityEventType = "ip_blocked"

	// Input validation events
	EventTypeValidationFailed    SecurityEventType = "validation_failed"
	EventTypeSQLInjectionAttempt SecurityEventType = "sql_injection_attempt"
	EventTypeXSSAttempt          SecurityEventType = "xss_attempt"
	EventTypeSuspiciousInput     SecurityEventType = "suspicious_input"

	// API events
	EventTypeAPIKeyCreated SecurityEventType = "api_key_created"
	EventTypeAPIKeyUsed    SecurityEventType = "api_key_used"
	EventTypeAPIKeyRevoked SecurityEventType = "api_key_revoked"

	// System events
	EventTypeSystemStartup  SecurityEventType = "system_startup"
	EventTypeSystemShutdown SecurityEventType = "system_shutdown"
	EventTypeConfigChange   SecurityEventType = "config_change"

	// Suspicious activity
	EventTypeSuspiciousUserAgent SecurityEventType = "suspicious_user_agent"
	EventTypeMultipleFailures    SecurityEventType = "multiple_failures"
	EventTypeUnusualActivity     SecurityEventType = "unusual_activity"
)

// SecuritySeverity represents the severity level of a security event
type SecuritySeverity string

const (
	SeverityLow      SecuritySeverity = "low"
	SeverityMedium   SecuritySeverity = "medium"
	SeverityHigh     SecuritySeverity = "high"
	SeverityCritical SecuritySeverity = "critical"
)

// SecurityAuditor handles security event logging and analysis
type SecurityAuditor struct {
	logger    *logrus.Logger
	events    chan SecurityEvent
	analyzers []SecurityAnalyzer
}

// SecurityAnalyzer interface for analyzing security events
type SecurityAnalyzer interface {
	Analyze(event SecurityEvent) (*SecurityAlert, error)
}

// SecurityAlert represents a security alert
type SecurityAlert struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	Severity    SecuritySeverity       `json:"severity"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	EventIDs    []string               `json:"event_ids"`
	RiskScore   int                    `json:"risk_score"`
	Actions     []string               `json:"actions"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// NewSecurityAuditor creates a new security auditor
func NewSecurityAuditor(logger *logrus.Logger) *SecurityAuditor {
	auditor := &SecurityAuditor{
		logger: logger,
		events: make(chan SecurityEvent, 1000),
		analyzers: []SecurityAnalyzer{
			NewBruteForceAnalyzer(),
			NewSuspiciousActivityAnalyzer(),
			NewRateLimitAnalyzer(),
		},
	}

	// Start event processing
	go auditor.processEvents()

	return auditor
}

// LogEvent logs a security event
func (sa *SecurityAuditor) LogEvent(event SecurityEvent) {
	// Set default values
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.Details == nil {
		event.Details = make(map[string]interface{})
	}

	// Calculate risk score if not set
	if event.RiskScore == 0 {
		event.RiskScore = sa.calculateRiskScore(event)
	}

	// Send to processing channel
	select {
	case sa.events <- event:
	default:
		// Channel is full, log directly
		sa.logEventDirectly(event)
	}
}

// LogLoginSuccess logs a successful login
func (sa *SecurityAuditor) LogLoginSuccess(userID uuid.UUID, ipAddress, userAgent, requestID string) {
	sa.LogEvent(SecurityEvent{
		EventType: EventTypeLoginSuccess,
		Severity:  SeverityLow,
		UserID:    &userID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		RequestID: requestID,
		Details: map[string]interface{}{
			"action": "user_login",
		},
	})
}

// LogLoginFailure logs a failed login attempt
func (sa *SecurityAuditor) LogLoginFailure(email, ipAddress, userAgent, requestID, reason string) {
	sa.LogEvent(SecurityEvent{
		EventType: EventTypeLoginFailure,
		Severity:  SeverityMedium,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		RequestID: requestID,
		Details: map[string]interface{}{
			"email":  email,
			"reason": reason,
		},
	})
}

// LogAccessDenied logs an access denied event
func (sa *SecurityAuditor) LogAccessDenied(userID *uuid.UUID, ipAddress, userAgent, requestID, endpoint, reason string) {
	sa.LogEvent(SecurityEvent{
		EventType: EventTypeAccessDenied,
		Severity:  SeverityMedium,
		UserID:    userID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		RequestID: requestID,
		Endpoint:  endpoint,
		Details: map[string]interface{}{
			"reason": reason,
		},
	})
}

// LogRateLimitExceeded logs a rate limit exceeded event
func (sa *SecurityAuditor) LogRateLimitExceeded(ipAddress, userAgent, requestID, endpoint string, limit int) {
	sa.LogEvent(SecurityEvent{
		EventType: EventTypeRateLimitExceeded,
		Severity:  SeverityHigh,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		RequestID: requestID,
		Endpoint:  endpoint,
		Details: map[string]interface{}{
			"limit": limit,
		},
	})
}

// LogDDoSDetected logs a DDoS detection event
func (sa *SecurityAuditor) LogDDoSDetected(ipAddress, userAgent, requestID string, requestCount int) {
	sa.LogEvent(SecurityEvent{
		EventType: EventTypeDDoSDetected,
		Severity:  SeverityCritical,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		RequestID: requestID,
		Blocked:   true,
		Details: map[string]interface{}{
			"request_count": requestCount,
		},
	})
}

// LogValidationFailed logs a validation failure
func (sa *SecurityAuditor) LogValidationFailed(ipAddress, userAgent, requestID, endpoint string, errors []string) {
	sa.LogEvent(SecurityEvent{
		EventType: EventTypeValidationFailed,
		Severity:  SeverityMedium,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		RequestID: requestID,
		Endpoint:  endpoint,
		Details: map[string]interface{}{
			"validation_errors": errors,
		},
	})
}

// LogSuspiciousInput logs a suspicious input attempt
func (sa *SecurityAuditor) LogSuspiciousInput(ipAddress, userAgent, requestID, endpoint, inputType, input string) {
	sa.LogEvent(SecurityEvent{
		EventType: EventTypeSuspiciousInput,
		Severity:  SeverityHigh,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		RequestID: requestID,
		Endpoint:  endpoint,
		Blocked:   true,
		Details: map[string]interface{}{
			"input_type": inputType,
			"input":      input,
		},
	})
}

// LogAPIKeyUsage logs API key usage
func (sa *SecurityAuditor) LogAPIKeyUsage(apiKeyID uuid.UUID, userID *uuid.UUID, ipAddress, userAgent, requestID, endpoint string) {
	sa.LogEvent(SecurityEvent{
		EventType: EventTypeAPIKeyUsed,
		Severity:  SeverityLow,
		UserID:    userID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		RequestID: requestID,
		Endpoint:  endpoint,
		Details: map[string]interface{}{
			"api_key_id": apiKeyID,
		},
	})
}

// processEvents processes security events
func (sa *SecurityAuditor) processEvents() {
	for event := range sa.events {
		// Log the event
		sa.logEventDirectly(event)

		// Analyze the event
		for _, analyzer := range sa.analyzers {
			if alert, err := analyzer.Analyze(event); err == nil && alert != nil {
				sa.logAlert(*alert)
			}
		}
	}
}

// logEventDirectly logs an event directly
func (sa *SecurityAuditor) logEventDirectly(event SecurityEvent) {
	// Create structured log entry
	entry := sa.logger.WithFields(logrus.Fields{
		"security_event": true,
		"event_id":       event.ID,
		"event_type":     event.EventType,
		"severity":       event.Severity,
		"ip_address":     event.IPAddress,
		"user_agent":     event.UserAgent,
		"request_id":     event.RequestID,
		"endpoint":       event.Endpoint,
		"method":         event.Method,
		"status":         event.Status,
		"risk_score":     event.RiskScore,
		"blocked":        event.Blocked,
	})

	if event.UserID != nil {
		entry = entry.WithField("user_id", event.UserID.String())
	}

	// Add details
	for key, value := range event.Details {
		entry = entry.WithField(key, value)
	}

	// Log with appropriate level
	switch event.Severity {
	case SeverityCritical:
		entry.Error("Security event: " + string(event.EventType))
	case SeverityHigh:
		entry.Warn("Security event: " + string(event.EventType))
	case SeverityMedium:
		entry.Info("Security event: " + string(event.EventType))
	case SeverityLow:
		entry.Debug("Security event: " + string(event.EventType))
	}
}

// logAlert logs a security alert
func (sa *SecurityAuditor) logAlert(alert SecurityAlert) {
	entry := sa.logger.WithFields(logrus.Fields{
		"security_alert": true,
		"alert_id":       alert.ID,
		"severity":       alert.Severity,
		"title":          alert.Title,
		"risk_score":     alert.RiskScore,
		"event_ids":      alert.EventIDs,
		"actions":        alert.Actions,
	})

	// Add metadata
	for key, value := range alert.Metadata {
		entry = entry.WithField(key, value)
	}

	// Log with appropriate level
	switch alert.Severity {
	case SeverityCritical:
		entry.Error("SECURITY ALERT: " + alert.Title)
	case SeverityHigh:
		entry.Warn("SECURITY ALERT: " + alert.Title)
	case SeverityMedium:
		entry.Info("SECURITY ALERT: " + alert.Title)
	case SeverityLow:
		entry.Debug("SECURITY ALERT: " + alert.Title)
	}
}

// calculateRiskScore calculates a risk score for an event
func (sa *SecurityAuditor) calculateRiskScore(event SecurityEvent) int {
	score := 0

	// Base score by event type
	switch event.EventType {
	case EventTypeLoginFailure:
		score += 20
	case EventTypeAccessDenied:
		score += 15
	case EventTypeRateLimitExceeded:
		score += 25
	case EventTypeDDoSDetected:
		score += 50
	case EventTypeSQLInjectionAttempt:
		score += 40
	case EventTypeXSSAttempt:
		score += 35
	case EventTypeSuspiciousInput:
		score += 30
	case EventTypeSuspiciousUserAgent:
		score += 20
	case EventTypeMultipleFailures:
		score += 35
	case EventTypeUnusualActivity:
		score += 25
	}

	// Adjust by severity
	switch event.Severity {
	case SeverityCritical:
		score += 30
	case SeverityHigh:
		score += 20
	case SeverityMedium:
		score += 10
	case SeverityLow:
		score += 5
	}

	// If blocked, increase score
	if event.Blocked {
		score += 15
	}

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return score
}

// GetSecurityStats returns security statistics
func (sa *SecurityAuditor) GetSecurityStats() map[string]interface{} {
	// This would typically query a database or cache
	// For now, return basic stats
	return map[string]interface{}{
		"total_events":     0,
		"blocked_requests": 0,
		"high_risk_events": 0,
		"active_threats":   0,
	}
}
