package updatesvc

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

type UpdateRequest struct {
	Context BecknContext   `json:"context"`
	Message UpdateMessage  `json:"message"`
}

// UpdateMessage carries a partial contract update.
type UpdateMessage struct {
	Contract json.RawMessage `json:"contract"`
}

type OnUpdateRequest struct {
	Context BecknContext     `json:"context"`
	Message OnUpdateMessage  `json:"message"`
}

type OnUpdateMessage struct {
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
