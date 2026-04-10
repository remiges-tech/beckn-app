package statussvc

import "encoding/json"

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

// StatusRequest is the inbound payload from the BAP: POST /beckn/status.
type StatusRequest struct {
	Context BecknContext    `json:"context"`
	Message StatusMessage   `json:"message"`
}

type StatusMessage struct {
	Contract ContractRef `json:"contract"`
}

// ContractRef carries only the id; the BPP looks up the full state.
type ContractRef struct {
	ID string `json:"id,omitempty"`
}

// OnStatusRequest is the outbound callback to the BAP.
type OnStatusRequest struct {
	Context BecknContext    `json:"context"`
	Message OnStatusMessage `json:"message"`
}

type OnStatusMessage struct {
	Contract json.RawMessage `json:"contract"`
}

type BecknACK struct {
	Context BecknContext `json:"context"`
	Message ACKMessage   `json:"message"`
}

type ACKMessage struct {
	ACK ACKStatus `json:"ack"`
}

type ACKStatus struct {
	Status string `json:"status"`
}
