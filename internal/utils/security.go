// ==========================
// internal/utils/security.go
// ==========================
package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
)

// Example usage in your B2C service:
// securityCredential, err := utils.EncryptInitiatorPassword(
//     s.config.InitiatorPassword,
//     "certs/production.cer",  // or "certs/sandbox.cer"
// )

// EncryptInitiatorPassword encrypts the initiator password using M-Pesa's public certificate
// This is required for B2C, B2B, and other sensitive M-Pesa API operations
func EncryptInitiatorPassword(initiatorPassword string, certificatePath string) (string, error) {
	// Read the certificate file
	certData, err := os.ReadFile(certificatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read certificate file: %w", err)
	}

	// Decode PEM block
	block, _ := pem.Decode(certData)
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block from certificate")
	}

	// Parse the certificate
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Extract public key
	publicKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("certificate does not contain RSA public key")
	}

	// Encrypt the password using RSA-OAEP
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, []byte(initiatorPassword))
	if err != nil {
		return "", fmt.Errorf("failed to encrypt password: %w", err)
	}

	// Encode to base64
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// EncryptInitiatorPasswordFromString encrypts using certificate content as string
// Useful when certificate is stored as environment variable or in database
func EncryptInitiatorPasswordFromString(initiatorPassword string, certificateContent string) (string, error) {
	// Decode PEM block
	block, _ := pem.Decode([]byte(certificateContent))
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block from certificate")
	}

	// Parse the certificate
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Extract public key
	publicKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("certificate does not contain RSA public key")
	}

	// Encrypt the password using RSA-PKCS1v15 (M-Pesa uses this, not OAEP)
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, []byte(initiatorPassword))
	if err != nil {
		return "", fmt.Errorf("failed to encrypt password: %w", err)
	}

	// Encode to base64
	return base64.StdEncoding.EncodeToString(encrypted), nil
}
