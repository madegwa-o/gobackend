// ==========================
// internal/services/b2c.go (FIXED)
// ==========================
package services

import (
	"awesomeProject/internal/config"
	"awesomeProject/internal/models"
	"awesomeProject/internal/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type B2CService struct {
	config      *config.Config
	authService *AuthService
}

func NewB2CService(cfg *config.Config, authService *AuthService) *B2CService {
	return &B2CService{
		config:      cfg,
		authService: authService,
	}
}

func (s *B2CService) InitiatePayment(req *models.B2CPaymentRequest) (*models.B2CPaymentResponse, error) {
	// Get access token
	accessToken, err := s.authService.GetAccessToken(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Encrypt initiator password
	securityCredential, err := utils.EncryptInitiatorPassword(
		s.config.InitiatorPassword,
		s.config.CertificatePath,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt password: %w", err)
	}

	// Prepare request payload - note "Occassion" typo matches M-Pesa API
	payload := map[string]interface{}{
		"OriginatorConversationID": req.OriginatorConversationID,
		"InitiatorName":            s.config.InitiatorName,
		"SecurityCredential":       securityCredential,
		"CommandID":                req.CommandID,
		"Amount":                   req.Amount,
		"PartyA":                   fmt.Sprintf("%d", s.config.BusinessShortCode),
		"PartyB":                   req.PhoneNumber,
		"Remarks":                  req.Remarks,
		"QueueTimeOutURL":          s.config.B2CTimeoutURL,
		"ResultURL":                s.config.B2CResultURL,
		"Occassion":                req.Occasion, // Typo in M-Pesa API
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	httpReq, err := http.NewRequest("POST", s.config.B2CURL(), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: time.Duration(s.config.APITimeout) * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Log response for debugging
	log.Printf("M-Pesa B2C Response: Status=%d, Body=%s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response - M-Pesa uses PascalCase
	var result struct {
		ConversationID           string `json:"ConversationID"`
		OriginatorConversationID string `json:"OriginatorConversationID"`
		ResponseCode             string `json:"ResponseCode"`
		ResponseDescription      string `json:"ResponseDescription"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &models.B2CPaymentResponse{
		ConversationID:           result.ConversationID,
		OriginatorConversationID: result.OriginatorConversationID,
		ResponseCode:             result.ResponseCode,
		ResponseDescription:      result.ResponseDescription,
	}, nil
}
