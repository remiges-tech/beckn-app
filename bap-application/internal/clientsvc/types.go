package clientsvc

import (
	"encoding/json"
	"time"
)

// ClientSelectRequest is the payload from the React frontend to initiate a Select.
type ClientSelectRequest struct {
	BppID     string          `json:"bpp_id" binding:"required"`
	BppURI    string          `json:"bpp_uri" binding:"required"`
	NetworkID string          `json:"network_id" binding:"required"`
	Items     json.RawMessage `json:"items" binding:"required"` // The selected items
}

// ClientInitRequest — frontend initiates Init.
type ClientInitRequest struct {
	TransactionID string          `json:"transaction_id" binding:"required"`
	BppID         string          `json:"bpp_id" binding:"required"`
	BppURI        string          `json:"bpp_uri" binding:"required"`
	NetworkID     string          `json:"network_id" binding:"required"`
	Billing       json.RawMessage `json:"billing" binding:"required"`
	Fulfillments  json.RawMessage `json:"fulfillments" binding:"required"`
}

// ClientConfirmRequest — frontend initiates Confirm.
type ClientConfirmRequest struct {
	TransactionID string `json:"transaction_id" binding:"required"`
	BppID         string `json:"bpp_id" binding:"required"`
	BppURI        string `json:"bpp_uri" binding:"required"`
	NetworkID     string `json:"network_id" binding:"required"`
}

// ClientStatusResponse returns the current state of a transaction/contract.
type ClientStatusResponse struct {
	TransactionID string           `json:"transaction_id"`
	Status        string           `json:"status"`
	Contract      json.RawMessage  `json:"contract,omitempty"` // Latest snapshot
	UpdatedAt     time.Time        `json:"updated_at"`
}

// BecknRequest is the base structure for outbound Beckn calls (select/init/confirm).
type BecknRequest struct {
	Context BecknContext    `json:"context"`
	Message json.RawMessage `json:"message"`
}

// BecknContext follows the Beckn Protocol v2.0 context schema.
// Field names are camelCase as required by the v2 spec.
type BecknContext struct {
	Version       string `json:"version"`
	Action        string `json:"action"`
	Timestamp     string `json:"timestamp"`
	MessageID     string `json:"messageId"`
	TransactionID string `json:"transactionId"`
	BapID         string `json:"bapId"`
	BapURI        string `json:"bapUri"`
	BppID         string `json:"bppId,omitempty"`
	BppURI        string `json:"bppUri,omitempty"`
	TTL           string `json:"ttl,omitempty"`
	NetworkID     string `json:"networkId"`
}

// ClientDiscoverRequest — carries the optional free-text search term from the frontend.
// Populated from the ?q query parameter on GET /api/v1/discover.
type ClientDiscoverRequest struct {
	TextSearch string // maps to intent.textSearch in the outbound Beckn request
}
