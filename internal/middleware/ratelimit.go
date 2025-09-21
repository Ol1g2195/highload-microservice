package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

type RateLimitMiddleware struct {
	limiter *limiter.Limiter
	logger  *logrus.Logger
}

type RateLimitConfig struct {
	Requests int           // Number of requests
	Duration time.Duration // Duration window
}

func NewRateLimitMiddleware(config RateLimitConfig, logger *logrus.Logger) *RateLimitMiddleware {
	// Create rate limiter with memory store
	store := memory.NewStore()

	// Create rate limit instance
	rate := limiter.Rate{
		Period: config.Duration,
		Limit:  int64(config.Requests),
	}

	instance := limiter.New(store, rate)

	return &RateLimitMiddleware{
		limiter: instance,
		logger:  logger,
	}
}

// RateLimit middleware that applies rate limiting to requests
func (m *RateLimitMiddleware) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP
		clientIP := c.ClientIP()

		// Create context for rate limiter
		ctx := context.Background()

		// Get rate limit info
		context, err := m.limiter.Get(ctx, clientIP)
		if err != nil {
			m.logger.Errorf("Rate limiter error: %v", err)
			// If rate limiter fails, allow request (fail open)
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

		// Check if rate limit exceeded
		if context.Reached {
			m.logger.Warnf("Rate limit exceeded for IP: %s", clientIP)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": fmt.Sprintf("Too many requests. Try again in %d seconds", context.Reset-time.Now().Unix()),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// StrictRateLimit middleware with stricter limits for sensitive endpoints
func (m *RateLimitMiddleware) StrictRateLimit() gin.HandlerFunc {
	// Create stricter rate limiter
	store := memory.NewStore()
	rate := limiter.Rate{
		Period: 1 * time.Minute, // 1 minute window
		Limit:  5,               // 5 requests per minute
	}

	strictLimiter := limiter.New(store, rate)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		ctx := context.Background()

		context, err := strictLimiter.Get(ctx, clientIP)
		if err != nil {
			m.logger.Errorf("Strict rate limiter error: %v", err)
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

		if context.Reached {
			m.logger.Warnf("Strict rate limit exceeded for IP: %s", clientIP)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests to sensitive endpoint. Try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuthRateLimit middleware for authentication endpoints
func (m *RateLimitMiddleware) AuthRateLimit() gin.HandlerFunc {
	// Very strict rate limiter for auth endpoints
	store := memory.NewStore()
	rate := limiter.Rate{
		Period: 15 * time.Minute, // 15 minute window
		Limit:  5,                // 5 attempts per 15 minutes
	}

	authLimiter := limiter.New(store, rate)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		ctx := context.Background()

		context, err := authLimiter.Get(ctx, clientIP)
		if err != nil {
			m.logger.Errorf("Auth rate limiter error: %v", err)
			c.Next()
			return
		}

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", strconv.FormatInt(context.Limit, 10))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(context.Remaining, 10))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(context.Reset, 10))

		if context.Reached {
			m.logger.Warnf("Auth rate limit exceeded for IP: %s", clientIP)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Authentication rate limit exceeded",
				"message": "Too many authentication attempts. Try again in 15 minutes.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
