package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"highload-microservice/internal/config"

	"golang.org/x/term"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	secretManager, err := config.NewSecretManager()
	if err != nil {
		fmt.Printf("Error creating secret manager: %v\n", err)
		os.Exit(1)
	}

	switch command {
	case "encrypt":
		if len(os.Args) < 3 {
			fmt.Println("Usage: secrets encrypt <value>")
			os.Exit(1)
		}
		encryptValue(secretManager, os.Args[2])
	case "decrypt":
		if len(os.Args) < 3 {
			fmt.Println("Usage: secrets decrypt <encrypted_value>")
			os.Exit(1)
		}
		decryptValue(secretManager, os.Args[2])
	case "set":
		if len(os.Args) < 3 {
			fmt.Println("Usage: secrets set <key>")
			os.Exit(1)
		}
		setSecret(secretManager, os.Args[2])
	case "validate":
		validateSecrets()
	case "generate-key":
		generateNewKey()
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Secrets Management Utility")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  secrets encrypt <value>     - Encrypt a value")
	fmt.Println("  secrets decrypt <value>     - Decrypt a value")
	fmt.Println("  secrets set <key>           - Set a secret interactively")
	fmt.Println("  secrets validate            - Validate current secrets")
	fmt.Println("  secrets generate-key        - Generate a new encryption key")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  secrets encrypt 'my-secret-password'")
	fmt.Println("  secrets set JWT_SECRET")
	fmt.Println("  secrets validate")
}

func encryptValue(secretManager *config.SecretManager, value string) {
	encrypted, err := secretManager.Encrypt(value)
	if err != nil {
		fmt.Printf("Error encrypting value: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Encrypted value: enc:%s\n", encrypted)
}

func decryptValue(secretManager *config.SecretManager, value string) {
	// Always trim prefix (idempotent)
	value = strings.TrimPrefix(value, "enc:")

	decrypted, err := secretManager.Decrypt(value)
	if err != nil {
		fmt.Printf("Error decrypting value: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Decrypted value: %s\n", decrypted)
}

func setSecret(secretManager *config.SecretManager, key string) {
	fmt.Printf("Enter value for %s: ", key)

	// Read password without echoing
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Printf("Error reading password: %v\n", err)
		os.Exit(1)
	}

	value := string(bytePassword)
	fmt.Println() // New line after password input

	if value == "" {
		fmt.Println("Empty value provided, skipping...")
		return
	}

	encrypted, err := secretManager.Encrypt(value)
	if err != nil {
		fmt.Printf("Error encrypting value: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Set this environment variable:\n")
	fmt.Printf("export %s=\"enc:%s\"\n", key, encrypted)
}

func validateSecrets() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	errors := config.ValidateSecrets(cfg)
	if len(errors) == 0 {
		fmt.Println("✅ All secrets are properly configured!")
		return
	}

	fmt.Println("❌ Security issues found:")
	for _, err := range errors {
		fmt.Printf("  - %s\n", err)
	}
	fmt.Println("")
	fmt.Println("Use 'secrets set <key>' to set secure values.")
}

func generateNewKey() {
	key, err := config.GenerateEncryptionKey()
	if err != nil {
		fmt.Printf("Error generating key: %v\n", err)
		os.Exit(1)
	}

	keyStr := config.Base64Encode(key)
	fmt.Printf("New encryption key: %s\n", keyStr)
	fmt.Println("")
	fmt.Println("Set this as your ENCRYPTION_KEY environment variable:")
	fmt.Printf("export ENCRYPTION_KEY=\"%s\"\n", keyStr)
}
