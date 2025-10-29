// ==========================
// internal/models/b2c.go (FIXED)
// ==========================
package models

// Client request model (snake_case for JSON)
type B2CPaymentRequest struct {
	PhoneNumber              string `json:"phone_number" binding:"required"`
	Amount                   int    `json:"amount" binding:"required,gt=0"`
	CommandID                string `json:"command_id" binding:"omitempty,oneof=SalaryPayment BusinessPayment PromotionPayment"`
	Remarks                  string `json:"remarks" binding:"max=100"`
	Occasion                 string `json:"occasion,omitempty" binding:"max=100"`
	OriginatorConversationID string `json:"originator_conversation_id,omitempty"`
}

type B2CPaymentResponse struct {
	ConversationID           string `json:"conversation_id"`
	OriginatorConversationID string `json:"originator_conversation_id"`
	ResponseCode             string `json:"response_code"`
	ResponseDescription      string `json:"response_description"`
}

// M-Pesa callback models (PascalCase to match M-Pesa's response)
type B2CResultParameter struct {
	Key   string      `json:"Key"`
	Value interface{} `json:"Value"`
}

type B2CResultParameters struct {
	ResultParameter []B2CResultParameter `json:"ResultParameter"`
}

type B2CReferenceItem struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

type B2CReferenceData struct {
	ReferenceItem B2CReferenceItem `json:"ReferenceItem"`
}

type B2CCallback struct {
	ResultType               int                  `json:"ResultType"`
	ResultCode               int                  `json:"ResultCode"`
	ResultDesc               string               `json:"ResultDesc"`
	OriginatorConversationID string               `json:"OriginatorConversationID"`
	ConversationID           string               `json:"ConversationID"`
	TransactionID            string               `json:"TransactionID"`
	ResultParameters         *B2CResultParameters `json:"ResultParameters,omitempty"`
	ReferenceData            *B2CReferenceData    `json:"ReferenceData,omitempty"`
}

type B2CResultRequest struct {
	Result B2CCallback `json:"Result"`
}

// Helper method to extract result parameters as map
func (cb *B2CCallback) GetResultParametersMap() map[string]interface{} {
	result := make(map[string]interface{})
	if cb.ResultParameters != nil {
		for _, param := range cb.ResultParameters.ResultParameter {
			result[param.Key] = param.Value
		}
	}
	return result
}
