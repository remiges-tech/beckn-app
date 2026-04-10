package supportsvc

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

type SupportRequest struct {
	Context BecknContext    `json:"context"`
	Message SupportMessage  `json:"message"`
}

type SupportMessage struct {
	Support SupportInput `json:"support"`
}

type SupportInput struct {
	// OrderID / ContractID this support request is for.
	OrderID    string          `json:"orderId,omitempty"`
	Descriptor SupportDescriptor `json:"descriptor,omitempty"`
	// Channels requested by the BAP (email, phone, chat, etc.).
	Channels   json.RawMessage `json:"channels,omitempty"`
}

type SupportDescriptor struct {
	Name      string `json:"name,omitempty"`
	ShortDesc string `json:"shortDesc,omitempty"`
}

type OnSupportRequest struct {
	Context BecknContext      `json:"context"`
	Message OnSupportMessage  `json:"message"`
}

type OnSupportMessage struct {
	Support OnSupportPayload `json:"support"`
}

// OnSupportPayload echoes back the ticket ID and the configured support channels.
type OnSupportPayload struct {
	TicketID string          `json:"ticketId"`
	Channels json.RawMessage `json:"channels"`
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
