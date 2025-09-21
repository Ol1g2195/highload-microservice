package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type DDoSProtection struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	logger   *logrus.Logger

	// Configuration
	maxRequests     int           // Maximum requests per window
	windowDuration  time.Duration // Time window
	blockDuration   time.Duration // How long to block IP
	cleanupInterval time.Duration // How often to cleanup old entries
}

type DDoSConfig struct {
	MaxRequests     int           // Maximum requests per window (default: 100)
	WindowDuration  time.Duration // Time window (default: 1 minute)
	BlockDuration   time.Duration // Block duration (default: 5 minutes)
	CleanupInterval time.Duration // Cleanup interval (default: 1 minute)
}

func NewDDoSProtection(config DDoSConfig, logger *logrus.Logger) *DDoSProtection {
	if config.MaxRequests == 0 {
		config.MaxRequests = 100
	}
	if config.WindowDuration == 0 {
		config.WindowDuration = 1 * time.Minute
	}
	if config.BlockDuration == 0 {
		config.BlockDuration = 5 * time.Minute
	}
	if config.CleanupInterval == 0 {
		config.CleanupInterval = 1 * time.Minute
	}

	ddos := &DDoSProtection{
		requests:        make(map[string][]time.Time),
		logger:          logger,
		maxRequests:     config.MaxRequests,
		windowDuration:  config.WindowDuration,
		blockDuration:   config.BlockDuration,
		cleanupInterval: config.CleanupInterval,
	}

	// Start cleanup goroutine
	go ddos.cleanup()

	return ddos
}

// Protect middleware that implements DDoS protection
func (d *DDoSProtection) Protect() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		// Check if IP is blocked
		if d.isBlocked(clientIP, now) {
			d.logger.Warnf("Blocked request from IP: %s (DDoS protection)", clientIP)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Request blocked",
				"message": "Your IP has been temporarily blocked due to suspicious activity.",
			})
			c.Abort()
			return
		}

		// Record request
		d.recordRequest(clientIP, now)

		// Check if IP should be blocked
		if d.shouldBlock(clientIP, now) {
			d.blockIP(clientIP, now)
			d.logger.Warnf("IP blocked due to DDoS: %s", clientIP)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Request blocked",
				"message": "Your IP has been temporarily blocked due to suspicious activity.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// isBlocked checks if an IP is currently blocked
func (d *DDoSProtection) isBlocked(ip string, now time.Time) bool {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	requests, exists := d.requests[ip]
	if !exists {
		return false
	}

	// Check if the last request was within block duration
	if len(requests) > 0 {
		lastRequest := requests[len(requests)-1]
		if now.Sub(lastRequest) < d.blockDuration {
			return true
		}
	}

	return false
}

// recordRequest records a request from an IP
func (d *DDoSProtection) recordRequest(ip string, now time.Time) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// Add current request
	d.requests[ip] = append(d.requests[ip], now)
}

// shouldBlock determines if an IP should be blocked
func (d *DDoSProtection) shouldBlock(ip string, now time.Time) bool {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	requests, exists := d.requests[ip]
	if !exists {
		return false
	}

	// Count requests within the window
	windowStart := now.Add(-d.windowDuration)
	count := 0

	for _, reqTime := range requests {
		if reqTime.After(windowStart) {
			count++
		}
	}

	return count > d.maxRequests
}

// blockIP blocks an IP by adding a special marker
func (d *DDoSProtection) blockIP(ip string, now time.Time) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// Add a special "blocked" timestamp
	d.requests[ip] = append(d.requests[ip], now)
}

// cleanup removes old entries to prevent memory leaks
func (d *DDoSProtection) cleanup() {
	ticker := time.NewTicker(d.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		d.mutex.Lock()
		now := time.Now()
		cutoff := now.Add(-d.blockDuration * 2) // Keep some history

		for ip, requests := range d.requests {
			// Remove old requests
			var newRequests []time.Time
			for _, reqTime := range requests {
				if reqTime.After(cutoff) {
					newRequests = append(newRequests, reqTime)
				}
			}

			if len(newRequests) == 0 {
				delete(d.requests, ip)
			} else {
				d.requests[ip] = newRequests
			}
		}
		d.mutex.Unlock()
	}
}

// GetStats returns current protection statistics
func (d *DDoSProtection) GetStats() map[string]interface{} {
	d.mutex.RLock()
	defer d.mutex.RUnlock()

	now := time.Now()
	windowStart := now.Add(-d.windowDuration)

	stats := map[string]interface{}{
		"total_ips":       len(d.requests),
		"max_requests":    d.maxRequests,
		"window_duration": d.windowDuration.String(),
		"block_duration":  d.blockDuration.String(),
		"active_requests": 0,
		"blocked_ips":     0,
	}

	activeRequests := 0
	blockedIPs := 0

	for ip, requests := range d.requests {
		// Count active requests
		for _, reqTime := range requests {
			if reqTime.After(windowStart) {
				activeRequests++
			}
		}

		// Check if IP is blocked
		if d.isBlocked(ip, now) {
			blockedIPs++
		}
	}

	stats["active_requests"] = activeRequests
	stats["blocked_ips"] = blockedIPs

	return stats
}
