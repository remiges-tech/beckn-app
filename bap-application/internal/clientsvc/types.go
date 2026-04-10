package clientsvc

import (
	"encoding/json"
	"time"
)

// SelectItem represents a single line in a select request.
// The frontend supplies the offer and provider details discovered earlier,
// including the BPP identity from the catalog so the request is routed to
// the correct BPP rather than the one hard-coded in server config.
type SelectItem struct {
	ResourceID      string          `json:"resource_id"`
	OfferID         string          `json:"offer_id"`
	OfferName       string          `json:"offer_name,omitempty"`
	Quantity        int             `json:"quantity"`
	ProviderID      string          `json:"provider_id,omitempty"`
	ProviderName    string          `json:"provider_name,omitempty"`
	OfferAttributes json.RawMessage `json:"offer_attributes,omitempty"`
	// BppID and BppURI are populated from the catalog's bppId/bppUri fields in the
	// discover response. They override the server-config defaults so the select is
	// routed to the BPP that actually owns the resource.
	BppID  string `json:"bpp_id,omitempty"`
	BppURI string `json:"bpp_uri,omitempty"`
}

// ClientSelectRequest is the payload from the React frontend to initiate a Select.
// BPP target is taken from the first item's BppID/BppURI (set from the catalog);
// server config values are used only as a fallback when those fields are empty.
type ClientSelectRequest struct {
	Items []SelectItem `json:"items" binding:"required"`
}

// ClientInitRequest — frontend initiates Init.
// Billing carries the structured buyer address/contact details; the BAP service
// patches them into the on_select contract snapshot before sending to the adapter.
type ClientInitRequest struct {
	TransactionID string          `json:"transaction_id" binding:"required"`
	Billing       json.RawMessage `json:"billing" binding:"required"`
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
