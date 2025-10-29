// ==========================
// internal/handlers/b2c.go (FIXED)
// ==========================
package handlers

import (
	"awesomeProject/internal/config"
	"awesomeProject/internal/models"
	"awesomeProject/internal/services"
	"awesomeProject/internal/utils"
	ws "awesomeProject/internal/websocket"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type B2CHandler struct {
	config     *config.Config
	b2cService *services.B2CService
	hub        *ws.Hub
}

func NewB2CHandler(cfg *config.Config, b2cService *services.B2CService, hub *ws.Hub) *B2CHandler {
	return &B2CHandler{
		config:     cfg,
		b2cService: b2cService,
		hub:        hub,
	}
}

func (h *B2CHandler) InitiatePayment(c *gin.Context) {
	var req models.B2CPaymentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:     "Invalid request",
			ErrorCode: "INVALID_REQUEST",
			Details:   map[string]interface{}{"error": err.Error()},
			Timestamp: time.Now(),
		})
		return
	}

	// Format phone number
	formattedPhone, err := utils.FormatPhoneNumber(req.PhoneNumber)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:     "Invalid phone number",
			ErrorCode: "INVALID_PHONE",
			Details:   map[string]interface{}{"error": err.Error()},
			Timestamp: time.Now(),
		})
		return
	}
	req.PhoneNumber = formattedPhone

	// Generate OriginatorConversationID if not provided
	if req.OriginatorConversationID == "" {
		req.OriginatorConversationID = uuid.New().String()
	}

	// Set default values (now that validation allows empty)
	if req.CommandID == "" {
		req.CommandID = "BusinessPayment"
	}
	if req.Remarks == "" {
		req.Remarks = "Payment"
	}

	// Initiate payment
	resp, err := h.b2cService.InitiatePayment(&req)
	if err != nil {
		log.Printf("B2C payment error: %v", err)
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:     "Failed to initiate payment",
			ErrorCode: "PAYMENT_FAILED",
			Details:   map[string]interface{}{"error": err.Error()},
			Timestamp: time.Now(),
		})
		return
	}

	// Broadcast initiation via WebSocket
	h.hub.BroadcastPaymentStatus(map[string]interface{}{
		"type":                       "b2c_initiated",
		"conversation_id":            resp.ConversationID,
		"originator_conversation_id": resp.OriginatorConversationID,
		"phone_number":               req.PhoneNumber,
		"amount":                     req.Amount,
		"command_id":                 req.CommandID,
		"timestamp":                  time.Now(),
	})

	c.JSON(http.StatusOK, models.SuccessResponse{
		Message: "Payment initiated successfully",
		Data: map[string]interface{}{
			"conversation_id":            resp.ConversationID,
			"originator_conversation_id": resp.OriginatorConversationID,
			"response_code":              resp.ResponseCode,
			"response_description":       resp.ResponseDescription,
		},
		Timestamp: time.Now(),
	})
}

func (h *B2CHandler) HandleCallback(c *gin.Context) {
	var callbackReq models.B2CResultRequest

	if err := c.ShouldBindJSON(&callbackReq); err != nil {
		log.Printf("B2C callback binding error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid callback data"})
		return
	}

	result := callbackReq.Result

	log.Printf("B2C Callback received - ConversationID: %s, ResultCode: %d, ResultDesc: %s",
		result.ConversationID, result.ResultCode, result.ResultDesc)

	// Parse result parameters if successful
	if result.ResultCode == 0 {
		h.processSuccessfulPayment(&result)
	} else {
		h.processFailedPayment(&result)
	}

	// Get parameters as map for broadcasting
	resultParams := result.GetResultParametersMap()

	// Broadcast callback via WebSocket
	h.hub.BroadcastPaymentStatus(map[string]interface{}{
		"type":                       "b2c_callback",
		"conversation_id":            result.ConversationID,
		"originator_conversation_id": result.OriginatorConversationID,
		"transaction_id":             result.TransactionID,
		"result_code":                result.ResultCode,
		"result_desc":                result.ResultDesc,
		"result_parameters":          resultParams,
		"timestamp":                  time.Now(),
	})

	c.JSON(http.StatusOK, gin.H{"ResultCode": 0, "ResultDesc": "Accepted"})
}

func (h *B2CHandler) HandleTimeout(c *gin.Context) {
	var callbackReq models.B2CResultRequest

	if err := c.ShouldBindJSON(&callbackReq); err != nil {
		log.Printf("B2C timeout binding error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid timeout data"})
		return
	}

	result := callbackReq.Result

	log.Printf("B2C Timeout - ConversationID: %s, ResultDesc: %s",
		result.ConversationID, result.ResultDesc)

	// Broadcast timeout via WebSocket
	h.hub.BroadcastPaymentStatus(map[string]interface{}{
		"type":                       "b2c_timeout",
		"conversation_id":            result.ConversationID,
		"originator_conversation_id": result.OriginatorConversationID,
		"result_desc":                result.ResultDesc,
		"timestamp":                  time.Now(),
	})

	c.JSON(http.StatusOK, gin.H{"ResultCode": 0, "ResultDesc": "Accepted"})
}

func (h *B2CHandler) processSuccessfulPayment(result *models.B2CCallback) {
	log.Printf("B2C Payment successful - TransactionID: %s", result.TransactionID)

	// Extract payment details using helper method
	details := result.GetResultParametersMap()
	log.Printf("Payment details: %+v", details)

	// Access specific values
	if transactionAmount, ok := details["TransactionAmount"]; ok {
		log.Printf("Transaction Amount: %v", transactionAmount)
	}
	if transactionReceipt, ok := details["TransactionReceipt"]; ok {
		log.Printf("Transaction Receipt: %v", transactionReceipt)
	}
	if recipientName, ok := details["ReceiverPartyPublicName"]; ok {
		log.Printf("Recipient: %v", recipientName)
	}
	if isRegistered, ok := details["B2CRecipientIsRegisteredCustomer"]; ok {
		log.Printf("Is Registered Customer: %v", isRegistered)
	}

	// Here you can:
	// - Update database records
	// - Send notifications
	// - Trigger webhooks
	// - Update transaction status
}

func (h *B2CHandler) processFailedPayment(result *models.B2CCallback) {
	log.Printf("B2C Payment failed - ResultCode: %d, ResultDesc: %s",
		result.ResultCode, result.ResultDesc)

	// Here you can:
	// - Update database with failure status
	// - Send failure notifications
	// - Log for analysis
	// - Trigger retry logic if applicable
}
