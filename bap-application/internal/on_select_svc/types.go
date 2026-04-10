package onselectsvc

import "encoding/json"

// ---------------------------------------------------------------------------
// Beckn context
// ---------------------------------------------------------------------------

type BecknContext struct {
	Version       string `json:"version"`
	Action        string `json:"action"`
	Timestamp     string `json:"timestamp"`
	MessageID     string `json:"messageId"`
	TransactionID string `json:"transactionId"`
	BapID         string `json:"bapId"`
	BapURI        string `json:"bapUri"`
	BppID         string `json:"bppId"`
	BppURI        string `json:"bppUri"`
	TTL           string `json:"ttl,omitempty"`
	NetworkID     string `json:"networkId"`
}

// ---------------------------------------------------------------------------
// INBOUND — BPP → BAP: POST /bap/receiver/on_select
// ---------------------------------------------------------------------------

type OnSelectRequest struct {
	Context BecknContext    `json:"context"`
	Message OnSelectMessage `json:"message"`
}

type OnSelectMessage struct {
	Contract json.RawMessage `json:"contract"` // store the full contract as-is
}

// ---------------------------------------------------------------------------
// Synchronous ACK returned to the BPP immediately
// ---------------------------------------------------------------------------

type BecknACK struct {
	Context BecknContext `json:"context"`
	Message ACKMessage   `json:"message"`
}

type ACKMessage struct {
	ACK ACKStatus `json:"ack"`
}

type ACKStatus struct {
	Status string `json:"status"` // "ACK" | "NACK"
}
