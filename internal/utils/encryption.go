package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ncobase/ncore/logging/logger"
	"golang.org/x/crypto/pbkdf2"
)

var (
	// ErrInvalidCiphertext is returned when the ciphertext is invalid
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
	// ErrInvalidKey is returned when the encryption key is invalid
	ErrInvalidKey = errors.New("invalid encryption key")
	// ErrEmptyPlaintext is returned when trying to encrypt empty data
	ErrEmptyPlaintext = errors.New("plaintext cannot be empty")
)

// EncryptionService provides encryption/decryption capabilities
type EncryptionService struct {
	masterKey       []byte
	keyVersion      int
	rotationEnabled bool
}

// EncryptionConfig holds encryption configuration
type EncryptionConfig struct {
	MasterKey       string
	KeyVersion      int
	RotationEnabled bool
}

// NewEncryptionService creates a new encryption service
func NewEncryptionService(config *EncryptionConfig) (*EncryptionService, error) {
	if config == nil {
		return nil, errors.New("encryption config is required")
	}

	// Derive master key from provided key or environment variable
	masterKeyStr := config.MasterKey
	if masterKeyStr == "" {
		masterKeyStr = os.Getenv("ENCRYPTION_MASTER_KEY")
	}

	if masterKeyStr == "" {
		return nil, errors.New("master key is required (set ENCRYPTION_MASTER_KEY environment variable)")
	}

	// Derive a 32-byte key using PBKDF2
	masterKey := deriveKey(masterKeyStr, "ncobase-encryption-salt", 32)

	keyVersion := config.KeyVersion
	if keyVersion == 0 {
		keyVersion = 1 // Default to version 1
	}

	return &EncryptionService{
		masterKey:       masterKey,
		keyVersion:      keyVersion,
		rotationEnabled: config.RotationEnabled,
	}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM
func (s *EncryptionService) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", ErrEmptyPlaintext
	}

	// Create cipher block
	block, err := aes.NewCipher(s.masterKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Use GCM mode for authenticated encryption
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the plaintext
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Prepend version and encode to base64
	versioned := fmt.Sprintf("v%d:%s", s.keyVersion, base64.StdEncoding.EncodeToString(ciphertext))

	return versioned, nil
}

// Decrypt decrypts ciphertext using AES-256-GCM
func (s *EncryptionService) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", ErrInvalidCiphertext
	}

	// Parse version prefix
	parts := strings.SplitN(ciphertext, ":", 2)
	if len(parts) != 2 {
		return "", ErrInvalidCiphertext
	}

	version := parts[0]
	encodedData := parts[1]

	// Validate version format
	if !strings.HasPrefix(version, "v") {
		return "", ErrInvalidCiphertext
	}

	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	// Create cipher block
	block, err := aes.NewCipher(s.masterKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Use GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrInvalidCiphertext
	}

	nonce, ciphertextData := data[:nonceSize], data[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertextData, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// EncryptMap encrypts sensitive fields in a map
func (s *EncryptionService) EncryptMap(data map[string]any, sensitiveFields []string) (map[string]any, error) {
	result := make(map[string]any)

	// Copy all fields
	for k, v := range data {
		result[k] = v
	}

	// Encrypt sensitive fields
	for _, field := range sensitiveFields {
		if val, exists := data[field]; exists && val != nil {
			strVal, ok := val.(string)
			if !ok {
				continue
			}

			encrypted, err := s.Encrypt(strVal)
			if err != nil {
				return nil, fmt.Errorf("failed to encrypt field %s: %w", field, err)
			}

			result[field] = encrypted
		}
	}

	return result, nil
}

// DecryptMap decrypts sensitive fields in a map
func (s *EncryptionService) DecryptMap(data map[string]any, sensitiveFields []string) (map[string]any, error) {
	result := make(map[string]any)

	// Copy all fields
	for k, v := range data {
		result[k] = v
	}

	// Decrypt sensitive fields
	for _, field := range sensitiveFields {
		if val, exists := data[field]; exists && val != nil {
			strVal, ok := val.(string)
			if !ok {
				continue
			}

			// Skip if not encrypted (doesn't have version prefix)
			if !strings.Contains(strVal, ":") || !strings.HasPrefix(strVal, "v") {
				continue
			}

			decrypted, err := s.Decrypt(strVal)
			if err != nil {
				logger.Warnf(nil, "Failed to decrypt field %s: %v", field, err)
				// Keep encrypted value if decryption fails
				continue
			}

			result[field] = decrypted
		}
	}

	return result, nil
}

// IsEncrypted checks if a string appears to be encrypted
func (s *EncryptionService) IsEncrypted(value string) bool {
	if value == "" {
		return false
	}

	// Check for version prefix pattern
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 {
		return false
	}

	return strings.HasPrefix(parts[0], "v")
}

// RotateKey re-encrypts data with a new key version
func (s *EncryptionService) RotateKey(ciphertext string, newKey []byte, newVersion int) (string, error) {
	if !s.rotationEnabled {
		return "", errors.New("key rotation is not enabled")
	}

	// Decrypt with current key
	plaintext, err := s.Decrypt(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt for rotation: %w", err)
	}

	// Create temporary service with new key
	tempService := &EncryptionService{
		masterKey:       newKey,
		keyVersion:      newVersion,
		rotationEnabled: s.rotationEnabled,
	}

	// Encrypt with new key
	newCiphertext, err := tempService.Encrypt(plaintext)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt with new key: %w", err)
	}

	return newCiphertext, nil
}

// GetKeyVersion returns the current key version
func (s *EncryptionService) GetKeyVersion() int {
	return s.keyVersion
}

// ValidateKey validates if a key can decrypt a ciphertext
func (s *EncryptionService) ValidateKey(ciphertext string) error {
	_, err := s.Decrypt(ciphertext)
	return err
}

// deriveKey derives an encryption key using PBKDF2
func deriveKey(password, salt string, keyLen int) []byte {
	return pbkdf2.Key([]byte(password), []byte(salt), 100000, keyLen, sha256.New)
}

// GenerateKey generates a secure random encryption key
func GenerateKey(length int) (string, error) {
	if length < 16 {
		return "", errors.New("key length must be at least 16 bytes")
	}

	key := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", fmt.Errorf("failed to generate key: %w", err)
	}

	return base64.StdEncoding.EncodeToString(key), nil
}

// SensitiveDataFields returns commonly sensitive field names
func SensitiveDataFields() []string {
	return []string{
		"password",
		"secret",
		"api_key",
		"apiKey",
		"api_secret",
		"apiSecret",
		"access_token",
		"accessToken",
		"refresh_token",
		"refreshToken",
		"private_key",
		"privateKey",
		"credential",
		"credentials",
		"auth_token",
		"authToken",
		"connection_string",
		"connectionString",
	}
}
