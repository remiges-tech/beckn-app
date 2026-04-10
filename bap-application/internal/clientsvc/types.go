package clientsvc

import (
	"encoding/json"
	"time"
)

// SelectItem represents a single line in a select request.
// The frontend supplies the offer and provider details discovered earlier.
type SelectItem struct {
	ResourceID      string          `json:"resource_id"`
	OfferID         string          `json:"offer_id"`
	OfferName       string          `json:"offer_name,omitempty"`
	Quantity        int             `json:"quantity"`
	ProviderID      string          `json:"provider_id,omitempty"`
	ProviderName    string          `json:"provider_name,omitempty"`
	OfferAttributes json.RawMessage `json:"offer_attributes,omitempty"`
}

// ClientSelectRequest is the payload from the React frontend to initiate a Select.
// BPP target (BppID, BppURI, NetworkID) is resolved from server config — not client-supplied.
type ClientSelectRequest struct {
	Items []SelectItem `json:"items" binding:"required"`
}

// ClientInitRequest — frontend initiates Init.
type ClientInitRequest struct {
	TransactionID string          `json:"transaction_id" binding:"required"`
	Billing       json.RawMessage `json:"billing" binding:"required"`
	Fulfillments  json.RawMessage `json:"fulfillments" binding:"required"`
}

// ClientConfirmRequest — frontend initiates Confirm.
type ClientConfirmRequest struct {
	TransactionID string `json:"transaction_id" binding:"required"`
}

// ClientStatusResponse returns the current state of a transaction/contract.
type ClientStatusResponse struct {
	TransactionID string          `json:"transaction_id"`
	Status        string          `json:"status"`
	Contract      json.RawMessage `json:"contract,omitempty"` // Latest snapshot
	UpdatedAt     time.Time       `json:"updated_at"`
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
