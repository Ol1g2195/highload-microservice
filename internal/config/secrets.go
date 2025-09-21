package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
)

// SecretManager handles secure storage and retrieval of secrets
type SecretManager struct {
	encryptionKey []byte
}

// NewSecretManager creates a new secret manager
func NewSecretManager() (*SecretManager, error) {
	key := getEnv("ENCRYPTION_KEY", "")
	if key == "" {
		// Generate a new key if none provided
		newKey, err := generateEncryptionKey()
		if err != nil {
			return nil, fmt.Errorf("failed to generate encryption key: %w", err)
		}
		key = base64.StdEncoding.EncodeToString(newKey)
		fmt.Printf("Generated new encryption key: %s\n", key)
		fmt.Println("IMPORTANT: Save this key securely and set ENCRYPTION_KEY environment variable")
	}

	keyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("invalid encryption key format: %w", err)
	}

	return &SecretManager{
		encryptionKey: keyBytes,
	}, nil
}

// Encrypt encrypts a plaintext string
func (sm *SecretManager) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(sm.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts an encrypted string
func (sm *SecretManager) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(sm.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// GetSecureEnv retrieves and decrypts a secure environment variable
func (sm *SecretManager) GetSecureEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	// Check if the value is encrypted (starts with "enc:")
	if strings.HasPrefix(value, "enc:") {
		encryptedValue := strings.TrimPrefix(value, "enc:")
		decrypted, err := sm.Decrypt(encryptedValue)
		if err != nil {
			fmt.Printf("Warning: Failed to decrypt %s: %v\n", key, err)
			return defaultValue
		}
		return decrypted
	}

	return value
}

// SetSecureEnv encrypts and sets an environment variable
func (sm *SecretManager) SetSecureEnv(key, value string) error {
	encrypted, err := sm.Encrypt(value)
	if err != nil {
		return err
	}

	encryptedValue := "enc:" + encrypted
	return os.Setenv(key, encryptedValue)
}

// generateEncryptionKey generates a new 32-byte encryption key
func generateEncryptionKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

// GenerateEncryptionKey is a public function to generate encryption keys
func GenerateEncryptionKey() ([]byte, error) {
	return generateEncryptionKey()
}

// Base64Encode encodes bytes to base64 string
func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// ValidateSecrets validates that all required secrets are present and valid
func ValidateSecrets(cfg *Config) []string {
	var errors []string

	// Check JWT secret
	if cfg.Auth.JWTSecret == "" || cfg.Auth.JWTSecret == "your-super-secret-jwt-key-change-in-production" {
		errors = append(errors, "JWT_SECRET must be set to a secure value")
	}

	// Check database password
	if cfg.Database.Password == "" || cfg.Database.Password == "postgres" {
		errors = append(errors, "DB_PASSWORD must be set to a secure value")
	}

	// Check Redis password (if required)
	if cfg.Redis.Password == "" && cfg.Redis.Host != "localhost" {
		errors = append(errors, "REDIS_PASSWORD should be set for production")
	}

	// Check TLS certificates
	if cfg.Server.UseTLS {
		if cfg.Server.TLSCert == "" || cfg.Server.TLSKey == "" {
			errors = append(errors, "TLS_CERT and TLS_KEY must be set when USE_TLS=true")
		}
	}

	return errors
}

// SanitizeConfig removes sensitive information from config for logging
func SanitizeConfig(cfg *Config) map[string]interface{} {
	return map[string]interface{}{
		"server": map[string]interface{}{
			"host":     cfg.Server.Host,
			"port":     cfg.Server.Port,
			"use_tls":  cfg.Server.UseTLS,
			"tls_cert": maskSensitive(cfg.Server.TLSCert),
			"tls_key":  maskSensitive(cfg.Server.TLSKey),
		},
		"database": map[string]interface{}{
			"host":     cfg.Database.Host,
			"port":     cfg.Database.Port,
			"user":     cfg.Database.User,
			"password": maskSensitive(cfg.Database.Password),
			"name":     cfg.Database.Name,
			"sslmode":  cfg.Database.SSLMode,
		},
		"redis": map[string]interface{}{
			"host":     cfg.Redis.Host,
			"port":     cfg.Redis.Port,
			"password": maskSensitive(cfg.Redis.Password),
			"db":       cfg.Redis.DB,
		},
		"kafka": map[string]interface{}{
			"brokers":  cfg.Kafka.Brokers,
			"topic":    cfg.Kafka.Topic,
			"group_id": cfg.Kafka.GroupID,
		},
		"auth": map[string]interface{}{
			"jwt_secret":         maskSensitive(cfg.Auth.JWTSecret),
			"jwt_expiration":     cfg.Auth.JWTExpiration,
			"refresh_expiration": cfg.Auth.RefreshExpiration,
			"api_key_length":     cfg.Auth.APIKeyLength,
		},
		"rate_limit": map[string]interface{}{
			"enabled":                  cfg.RateLimit.Enabled,
			"requests_per_minute":      cfg.RateLimit.RequestsPerMinute,
			"burst_size":               cfg.RateLimit.BurstSize,
			"auth_requests_per_minute": cfg.RateLimit.AuthRequestsPerMinute,
			"auth_burst_size":          cfg.RateLimit.AuthBurstSize,
		},
		"log_level": cfg.LogLevel,
	}
}

// maskSensitive masks sensitive values for logging
func maskSensitive(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 8 {
		return "***"
	}
	return value[:4] + "***" + value[len(value)-4:]
}
