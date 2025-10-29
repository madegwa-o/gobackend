// ==========================
// internal/utils/utils.go
// ==========================
package utils

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"time"
)

func GeneratePassword(businessShortCode, passkey, timestamp string) string {
	data := fmt.Sprintf("%s%s%s", businessShortCode, passkey, timestamp)
	return base64.StdEncoding.EncodeToString([]byte(data))
}

func GetTimestamp() string {
	return time.Now().Format("20060102150405")
}

func FormatPhoneNumber(phoneNumber string) (string, error) {
	// Remove non-digit characters except '+'
	re := regexp.MustCompile(`[^\d+]`)
	cleaned := re.ReplaceAllString(phoneNumber, "")

	// Remove leading '+'
	if len(cleaned) > 0 && cleaned[0] == '+' {
		cleaned = cleaned[1:]
	}

	// Handle different formats
	if len(cleaned) > 0 && cleaned[0] == '0' {
		cleaned = "254" + cleaned[1:]
	} else if len(cleaned) >= 9 && !regexp.MustCompile(`^254`).MatchString(cleaned) {
		cleaned = "254" + cleaned
	}

	// Validate format
	if !regexp.MustCompile(`^254\d{9}$`).MatchString(cleaned) {
		return "", fmt.Errorf("invalid phone number format: %s. Expected format: 254XXXXXXXXX", phoneNumber)
	}

	return cleaned, nil
}
