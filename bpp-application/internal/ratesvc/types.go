package ratesvc

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

type RateRequest struct {
	Context BecknContext `json:"context"`
	Message RateMessage  `json:"message"`
}

type RateMessage struct {
	RatingInputs []RatingInput `json:"ratingInputs"`
}

type RatingInput struct {
	// ID of the entity being rated (contract, resource, provider, etc.)
	ID               string          `json:"id"`
	RatingCategory   string          `json:"ratingCategory,omitempty"`
	Descriptor       json.RawMessage `json:"descriptor,omitempty"`
	Range            json.RawMessage `json:"range"`
	FeedbackForm     json.RawMessage `json:"feedbackForm,omitempty"`
	FeedbackFormSubmission json.RawMessage `json:"feedbackFormSubmission,omitempty"`
}

type OnRateRequest struct {
	Context BecknContext  `json:"context"`
	Message OnRateMessage `json:"message"`
}

type OnRateMessage struct {
	Ratings []RatingSummary `json:"ratings"`
}

type RatingSummary struct {
	ID    string          `json:"id"`
	Range json.RawMessage `json:"range"`
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
