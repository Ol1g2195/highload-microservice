package security

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// BruteForceAnalyzer detects brute force attacks
type BruteForceAnalyzer struct {
	failedLogins map[string][]time.Time
	mu           sync.RWMutex
}

// NewBruteForceAnalyzer creates a new brute force analyzer
func NewBruteForceAnalyzer() *BruteForceAnalyzer {
	return &BruteForceAnalyzer{
		failedLogins: make(map[string][]time.Time),
	}
}

// Analyze analyzes events for brute force patterns
func (bfa *BruteForceAnalyzer) Analyze(event SecurityEvent) (*SecurityAlert, error) {
	if event.EventType != EventTypeLoginFailure {
		return nil, nil
	}

	bfa.mu.Lock()
	defer bfa.mu.Unlock()

	// Add failed login timestamp
	bfa.failedLogins[event.IPAddress] = append(bfa.failedLogins[event.IPAddress], event.Timestamp)

	// Clean old entries (older than 15 minutes)
	cutoff := time.Now().Add(-15 * time.Minute)
	var recentFailures []time.Time
	for _, timestamp := range bfa.failedLogins[event.IPAddress] {
		if timestamp.After(cutoff) {
			recentFailures = append(recentFailures, timestamp)
		}
	}
	bfa.failedLogins[event.IPAddress] = recentFailures

	// Check for brute force pattern
	if len(recentFailures) >= 5 {
		// Create alert
		alert := &SecurityAlert{
			ID:        uuid.New().String(),
			Timestamp: time.Now(),
			Severity:  SeverityHigh,
			Title:     "Brute Force Attack Detected",
			Description: fmt.Sprintf("IP %s has made %d failed login attempts in the last 15 minutes",
				event.IPAddress, len(recentFailures)),
			EventIDs:  []string{event.ID},
			RiskScore: 75,
			Actions: []string{
				"Consider blocking IP address",
				"Increase rate limiting for this IP",
				"Monitor for additional suspicious activity",
			},
			Metadata: map[string]interface{}{
				"ip_address":    event.IPAddress,
				"failure_count": len(recentFailures),
				"time_window":   "15 minutes",
				"attack_type":   "brute_force",
			},
		}

		// Clear the failed logins for this IP to avoid spam
		delete(bfa.failedLogins, event.IPAddress)

		return alert, nil
	}

	return nil, nil
}

// SuspiciousActivityAnalyzer detects suspicious activity patterns
type SuspiciousActivityAnalyzer struct {
	userActivity map[string][]SecurityEvent
	mu           sync.RWMutex
}

// NewSuspiciousActivityAnalyzer creates a new suspicious activity analyzer
func NewSuspiciousActivityAnalyzer() *SuspiciousActivityAnalyzer {
	return &SuspiciousActivityAnalyzer{
		userActivity: make(map[string][]SecurityEvent),
	}
}

// Analyze analyzes events for suspicious activity patterns
func (saa *SuspiciousActivityAnalyzer) Analyze(event SecurityEvent) (*SecurityAlert, error) {
	// Track user activity
	userKey := event.IPAddress
	if event.UserID != nil {
		userKey = event.UserID.String()
	}

	saa.mu.Lock()
	defer saa.mu.Unlock()

	// Add event to user activity
	saa.userActivity[userKey] = append(saa.userActivity[userKey], event)

	// Clean old events (older than 1 hour)
	cutoff := time.Now().Add(-1 * time.Hour)
	var recentEvents []SecurityEvent
	for _, e := range saa.userActivity[userKey] {
		if e.Timestamp.After(cutoff) {
			recentEvents = append(recentEvents, e)
		}
	}
	saa.userActivity[userKey] = recentEvents

	// Check for suspicious patterns
	if len(recentEvents) >= 10 {
		// Count different types of events
		eventTypes := make(map[SecurityEventType]int)
		blockedCount := 0
		highRiskCount := 0

		for _, e := range recentEvents {
			eventTypes[e.EventType]++
			if e.Blocked {
				blockedCount++
			}
			if e.RiskScore > 50 {
				highRiskCount++
			}
		}

		// Create alert if suspicious
		if blockedCount > 3 || highRiskCount > 5 {
			var eventIDs []string
			for _, e := range recentEvents {
				eventIDs = append(eventIDs, e.ID)
			}

			alert := &SecurityAlert{
				ID:        uuid.New().String(),
				Timestamp: time.Now(),
				Severity:  SeverityMedium,
				Title:     "Suspicious Activity Detected",
				Description: fmt.Sprintf("User/IP %s has shown suspicious activity with %d events, %d blocked, %d high-risk",
					userKey, len(recentEvents), blockedCount, highRiskCount),
				EventIDs:  eventIDs,
				RiskScore: 60,
				Actions: []string{
					"Review user activity",
					"Consider additional monitoring",
					"Check for account compromise",
				},
				Metadata: map[string]interface{}{
					"user_key":        userKey,
					"total_events":    len(recentEvents),
					"blocked_count":   blockedCount,
					"high_risk_count": highRiskCount,
					"event_types":     eventTypes,
					"time_window":     "1 hour",
				},
			}

			return alert, nil
		}
	}

	return nil, nil
}

// RateLimitAnalyzer analyzes rate limiting events
type RateLimitAnalyzer struct {
	rateLimitEvents map[string][]time.Time
	mu              sync.RWMutex
}

// NewRateLimitAnalyzer creates a new rate limit analyzer
func NewRateLimitAnalyzer() *RateLimitAnalyzer {
	return &RateLimitAnalyzer{
		rateLimitEvents: make(map[string][]time.Time),
	}
}

// Analyze analyzes rate limiting events
func (rla *RateLimitAnalyzer) Analyze(event SecurityEvent) (*SecurityAlert, error) {
	if event.EventType != EventTypeRateLimitExceeded && event.EventType != EventTypeDDoSDetected {
		return nil, nil
	}

	rla.mu.Lock()
	defer rla.mu.Unlock()

	// Add rate limit event
	rla.rateLimitEvents[event.IPAddress] = append(rla.rateLimitEvents[event.IPAddress], event.Timestamp)

	// Clean old entries (older than 1 hour)
	cutoff := time.Now().Add(-1 * time.Hour)
	var recentEvents []time.Time
	for _, timestamp := range rla.rateLimitEvents[event.IPAddress] {
		if timestamp.After(cutoff) {
			recentEvents = append(recentEvents, timestamp)
		}
	}
	rla.rateLimitEvents[event.IPAddress] = recentEvents

	// Check for persistent rate limiting
	if len(recentEvents) >= 10 {
		alert := &SecurityAlert{
			ID:        uuid.New().String(),
			Timestamp: time.Now(),
			Severity:  SeverityHigh,
			Title:     "Persistent Rate Limiting",
			Description: fmt.Sprintf("IP %s has exceeded rate limits %d times in the last hour",
				event.IPAddress, len(recentEvents)),
			EventIDs:  []string{event.ID},
			RiskScore: 70,
			Actions: []string{
				"Consider permanent IP blocking",
				"Increase rate limiting restrictions",
				"Monitor for DDoS attack patterns",
			},
			Metadata: map[string]interface{}{
				"ip_address":      event.IPAddress,
				"violation_count": len(recentEvents),
				"time_window":     "1 hour",
				"attack_type":     "rate_limit_abuse",
			},
		}

		// Clear the events for this IP to avoid spam
		delete(rla.rateLimitEvents, event.IPAddress)

		return alert, nil
	}

	return nil, nil
}

// SecurityMetrics tracks security metrics
type SecurityMetrics struct {
	TotalEvents          int64 `json:"total_events"`
	BlockedRequests      int64 `json:"blocked_requests"`
	HighRiskEvents       int64 `json:"high_risk_events"`
	ActiveThreats        int64 `json:"active_threats"`
	LoginFailures        int64 `json:"login_failures"`
	AccessDenied         int64 `json:"access_denied"`
	RateLimitHits        int64 `json:"rate_limit_hits"`
	DDoSAttempts         int64 `json:"ddos_attempts"`
	SQLInjectionAttempts int64 `json:"sql_injection_attempts"`
	XSSAttempts          int64 `json:"xss_attempts"`
	mu                   sync.RWMutex
}

// NewSecurityMetrics creates new security metrics
func NewSecurityMetrics() *SecurityMetrics {
	return &SecurityMetrics{}
}

// IncrementEvent increments event counter
func (sm *SecurityMetrics) IncrementEvent(eventType SecurityEventType, blocked bool, riskScore int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.TotalEvents++

	if blocked {
		sm.BlockedRequests++
	}

	if riskScore > 50 {
		sm.HighRiskEvents++
	}

	switch eventType {
	case EventTypeLoginFailure:
		sm.LoginFailures++
	case EventTypeAccessDenied:
		sm.AccessDenied++
	case EventTypeRateLimitExceeded:
		sm.RateLimitHits++
	case EventTypeDDoSDetected:
		sm.DDoSAttempts++
	case EventTypeSQLInjectionAttempt:
		sm.SQLInjectionAttempts++
	case EventTypeXSSAttempt:
		sm.XSSAttempts++
	}
}

// GetMetrics returns current metrics
func (sm *SecurityMetrics) GetMetrics() SecurityMetrics {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return SecurityMetrics{
		TotalEvents:          sm.TotalEvents,
		BlockedRequests:      sm.BlockedRequests,
		HighRiskEvents:       sm.HighRiskEvents,
		ActiveThreats:        sm.ActiveThreats,
		LoginFailures:        sm.LoginFailures,
		AccessDenied:         sm.AccessDenied,
		RateLimitHits:        sm.RateLimitHits,
		DDoSAttempts:         sm.DDoSAttempts,
		SQLInjectionAttempts: sm.SQLInjectionAttempts,
		XSSAttempts:          sm.XSSAttempts,
	}
}

// Reset resets all metrics
func (sm *SecurityMetrics) Reset() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.TotalEvents = 0
	sm.BlockedRequests = 0
	sm.HighRiskEvents = 0
	sm.ActiveThreats = 0
	sm.LoginFailures = 0
	sm.AccessDenied = 0
	sm.RateLimitHits = 0
	sm.DDoSAttempts = 0
	sm.SQLInjectionAttempts = 0
	sm.XSSAttempts = 0
}
