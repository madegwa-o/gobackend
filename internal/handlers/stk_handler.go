// ==========================
// internal/handlers/stk.go
// ==========================
package handlers

import (
	"awesomeProject/internal/config"
	"awesomeProject/internal/models"
	"awesomeProject/internal/services"
	"awesomeProject/internal/utils"
	"awesomeProject/internal/websocket"
	"log"

	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type STKHandler struct {
	config      *config.Config
	authService *services.AuthService
	hub         *websocket.Hub
}

func NewSTKHandler(cfg *config.Config, authSvc *services.AuthService, hub *websocket.Hub) *STKHandler {
	return &STKHandler{
		config:      cfg,
		authService: authSvc,
		hub:         hub,
	}
}

func (h *STKHandler) InitiateSTKPush(c *gin.Context) {
	var req models.STKPushRequest

	// Log raw request body for debugging
	rawBody, _ := c.GetRawData()
	log.Printf("📥 Received STK Push Request - Raw Body: %s", string(rawBody))

	// Rebind the body since we read it
	c.Request.Body = io.NopCloser(bytes.NewBuffer(rawBody))

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ Request binding failed: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:     "Invalid request",
			Details:   map[string]interface{}{"validation": err.Error()},
			Timestamp: time.Now(),
		})
		return
	}

	// Log parsed request details
	log.Printf("📋 Parsed STK Request - Phone: %s, Amount: %d, Account: %s, Desc: %s",
		req.PhoneNumber, req.Amount, req.AccountReference, req.TransactionDesc)

	accessToken, err := h.authService.GetAccessToken(false)
	if err != nil {
		log.Printf("❌ Failed to get access token: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:     "Failed to get access token",
			Details:   map[string]interface{}{"error": err.Error()},
			Timestamp: time.Now(),
		})
		return
	}

	log.Printf("🔑 Access token obtained successfully")

	timestamp := utils.GetTimestamp()
	password := utils.GeneratePassword(
		strconv.Itoa(h.config.BusinessShortCode),
		h.config.Passkey,
		timestamp,
	)

	phoneNumber, err := utils.FormatPhoneNumber(req.PhoneNumber)
	if err != nil {
		log.Printf("❌ Phone number formatting failed: %v", err)
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:     err.Error(),
			Timestamp: time.Now(),
		})
		return
	}

	log.Printf("📱 Formatted phone number: %s", phoneNumber)

	payload := map[string]interface{}{
		"BusinessShortCode": h.config.BusinessShortCode,
		"Password":          password,
		"Timestamp":         timestamp,
		"TransactionType":   "CustomerPayBillOnline",
		"Amount":            req.Amount,
		"PartyA":            phoneNumber,
		"PartyB":            h.config.BusinessShortCode,
		"PhoneNumber":       phoneNumber,
		"CallBackURL":       h.config.STKCallbackURL,
		"AccountReference":  req.AccountReference,
		"TransactionDesc":   req.TransactionDesc,
	}

	// Log the payload being sent to Safaricom (excluding password for security)
	logPayload := make(map[string]interface{})
	for k, v := range payload {
		if k != "Password" {
			logPayload[k] = v
		} else {
			logPayload[k] = "***REDACTED***"
		}
	}
	logPayloadJSON, _ := json.MarshalIndent(logPayload, "", "  ")
	log.Printf("📤 Sending STK Push to Safaricom:\n%s", string(logPayloadJSON))

	jsonPayload, _ := json.Marshal(payload)
	httpReq, _ := http.NewRequest("POST", h.config.STKPushURL(), bytes.NewBuffer(jsonPayload))
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)
	httpReq.Header.Set("Content-Type", "application/json")

	log.Printf("🌐 M-Pesa API URL: %s", h.config.STKPushURL())

	client := &http.Client{Timeout: 30 * time.Second}
	startTime := time.Now()
	resp, err := client.Do(httpReq)
	duration := time.Since(startTime)

	if err != nil {
		log.Printf("❌ HTTP request failed after %v: %v", duration, err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:     "Failed to initiate STK push",
			Details:   map[string]interface{}{"error": err.Error()},
			Timestamp: time.Now(),
		})
		return
	}
	defer resp.Body.Close()

	log.Printf("⏱️  M-Pesa API responded in %v with status: %d", duration, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	log.Printf("📨 M-Pesa API Response: %s", string(body))

	if resp.StatusCode != http.StatusOK {
		var errorData map[string]interface{}
		json.Unmarshal(body, &errorData)
		log.Printf("❌ STK Push failed - Status: %d, Response: %v", resp.StatusCode, errorData)
		c.JSON(resp.StatusCode, models.ErrorResponse{
			Error:     "STK push failed",
			ErrorCode: fmt.Sprintf("%v", errorData["errorCode"]),
			Details:   errorData,
			Timestamp: time.Now(),
		})
		return
	}

	var result models.STKPushResponse
	json.Unmarshal(body, &result)

	log.Printf("✅ STK Push initiated successfully - CheckoutRequestID: %s, MerchantRequestID: %s",
		result.CheckoutRequestID, result.MerchantRequestID)

	c.JSON(http.StatusOK, result)
}

func (h *STKHandler) STKPushCallback(c *gin.Context) {
	var req models.STKCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, models.SuccessResponse{
			Message:   "Callback received",
			Data:      map[string]interface{}{"error": err.Error()},
			Timestamp: time.Now(),
		})
		return
	}

	callback := req.Body.STKCallback

	// Broadcast to WebSocket clients
	h.hub.BroadcastPaymentStatus(map[string]interface{}{
		"type": "stk_callback",
		"data": callback,
	})

	if callback.ResultCode == 0 {
		fmt.Printf("✓ STK Push successful: %s\n", callback.CheckoutRequestID)
	} else {
		fmt.Printf("✗ STK Push failed: %s\n", callback.ResultDesc)
	}

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message:   "Callback received successfully",
		Timestamp: time.Now(),
	})
}
