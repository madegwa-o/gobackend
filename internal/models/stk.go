// ==========================
// internal/models/stk.go
// ==========================
package models

import "time"

type STKPushRequest struct {
	PhoneNumber      string `json:"phone_number" binding:"required"`
	Amount           int    `json:"amount" binding:"required,gt=0"`
	AccountReference string `json:"account_reference" binding:"required,max=12"`
	TransactionDesc  string `json:"transaction_desc" binding:"required,max=13"`
}

type STKPushResponse struct {
	MerchantRequestID   string `json:"MerchantRequestID"`   // Changed from merchant_request_id
	CheckoutRequestID   string `json:"CheckoutRequestID"`   // Changed from checkout_request_id
	ResponseCode        string `json:"ResponseCode"`        // Changed from response_code
	ResponseDescription string `json:"ResponseDescription"` // Changed from response_description
	CustomerMessage     string `json:"CustomerMessage"`     // Changed from customer_message
}

type CallbackMetadataItem struct {
	Name  string      `json:"Name"`
	Value interface{} `json:"Value,omitempty"`
}

type CallbackMetadata struct {
	Item []CallbackMetadataItem `json:"Item"`
}

type STKCallback struct {
	MerchantRequestID string            `json:"MerchantRequestID"`
	CheckoutRequestID string            `json:"CheckoutRequestID"`
	ResultCode        int               `json:"ResultCode"`
	ResultDesc        string            `json:"ResultDesc"`
	CallbackMetadata  *CallbackMetadata `json:"CallbackMetadata,omitempty"`
}

type STKPushCallback struct {
	MerchantRequestID string                 `json:"merchant_request_id"`
	CheckoutRequestID string                 `json:"checkout_request_id"`
	ResultCode        int                    `json:"result_code"`
	ResultDesc        string                 `json:"result_desc"`
	CallbackMetadata  map[string]interface{} `json:"callback_metadata,omitempty"`
}

type STKCallbackRequest struct {
	Body struct {
		STKCallback STKCallback `json:"stkCallback"`
	} `json:"Body"`
}

type SuccessResponse struct {
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

type ErrorResponse struct {
	Error     string                 `json:"error"`
	ErrorCode string                 `json:"error_code,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}
