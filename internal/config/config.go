package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	Kafka     KafkaConfig
	Auth      AuthConfig
	RateLimit RateLimitConfig
	Security  SecurityConfig
	LogLevel  string
}

type ServerConfig struct {
	Host    string
	Port    string
	TLSCert string
	TLSKey  string
	UseTLS  bool
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type KafkaConfig struct {
	Brokers []string
	Topic   string
	GroupID string
}

type AuthConfig struct {
	JWTSecret         string
	JWTExpiration     int // in hours
	RefreshExpiration int // in days
	APIKeyLength      int
}

type RateLimitConfig struct {
	Enabled               bool
	RequestsPerMinute     int
	BurstSize             int
	AuthRequestsPerMinute int
	AuthBurstSize         int
}

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

func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Initialize secret manager
	secretManager, err := NewSecretManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize secret manager: %w", err)
	}

	config := &Config{
		Server: ServerConfig{
			Host:    getEnv("SERVER_HOST", "0.0.0.0"),
			Port:    getEnv("SERVER_PORT", "8080"),
			TLSCert: getEnv("TLS_CERT", "certs/server.crt"),
			TLSKey:  getEnv("TLS_KEY", "certs/server.key"),
			UseTLS:  getEnvAsBool("USE_TLS", false),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: secretManager.GetSecureEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "highload_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: secretManager.GetSecureEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Kafka: KafkaConfig{
			Brokers: []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
			Topic:   getEnv("KAFKA_TOPIC", "user-events"),
			GroupID: getEnv("KAFKA_GROUP_ID", "highload-service"),
		},
		Auth: AuthConfig{
			JWTSecret:         secretManager.GetSecureEnv("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"),
			JWTExpiration:     getEnvAsInt("JWT_EXPIRATION_HOURS", 24),
			RefreshExpiration: getEnvAsInt("REFRESH_EXPIRATION_DAYS", 7),
			APIKeyLength:      getEnvAsInt("API_KEY_LENGTH", 32),
		},
		RateLimit: RateLimitConfig{
			Enabled:               getEnvAsBool("RATE_LIMIT_ENABLED", true),
			RequestsPerMinute:     getEnvAsInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 60),
			BurstSize:             getEnvAsInt("RATE_LIMIT_BURST_SIZE", 10),
			AuthRequestsPerMinute: getEnvAsInt("RATE_LIMIT_AUTH_REQUESTS_PER_MINUTE", 5),
			AuthBurstSize:         getEnvAsInt("RATE_LIMIT_AUTH_BURST_SIZE", 2),
		},
		Security: SecurityConfig{
			AllowedOrigins:        getEnvAsStringSlice("CORS_ALLOWED_ORIGINS", []string{"https://localhost:3000", "https://127.0.0.1:3000"}),
			AllowedMethods:        getEnvAsStringSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"}),
			AllowedHeaders:        getEnvAsStringSlice("CORS_ALLOWED_HEADERS", []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-Request-ID", "X-API-Key"}),
			ExposedHeaders:        getEnvAsStringSlice("CORS_EXPOSED_HEADERS", []string{"X-Request-ID", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset"}),
			AllowCredentials:      getEnvAsBool("CORS_ALLOW_CREDENTIALS", true),
			MaxAge:                getEnvAsInt("CORS_MAX_AGE", 86400),
			ContentTypeNosniff:    getEnvAsBool("SECURITY_CONTENT_TYPE_NOSNIFF", true),
			FrameDeny:             getEnvAsBool("SECURITY_FRAME_DENY", true),
			XSSProtection:         getEnvAsBool("SECURITY_XSS_PROTECTION", true),
			ReferrerPolicy:        getEnv("SECURITY_REFERRER_POLICY", "strict-origin-when-cross-origin"),
			PermissionsPolicy:     getEnv("SECURITY_PERMISSIONS_POLICY", "geolocation=(), microphone=(), camera=()"),
			ContentSecurityPolicy: getEnv("SECURITY_CSP", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self';"),
		},
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
