package cancelsvc

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

type CancelRequest struct {
	Context BecknContext   `json:"context"`
	Message CancelMessage  `json:"message"`
}

type CancelMessage struct {
	Contract ContractRef `json:"contract"`
}

type ContractRef struct {
	ID string `json:"id,omitempty"`
}

type OnCancelRequest struct {
	Context BecknContext     `json:"context"`
	Message OnCancelMessage  `json:"message"`
}

type OnCancelMessage struct {
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
