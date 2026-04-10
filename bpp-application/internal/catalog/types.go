package catalog

import "encoding/json"

// ProviderPublishRequest is what a provider sends to POST /v1/catalog/publish.
// Beckn context fields (bppId, networkId, etc.) are intentionally absent —
// the BPP injects them automatically from its own config before forwarding to the CDS.
type ProviderPublishRequest struct {
	Catalogs []Catalog `json:"catalogs" binding:"required,min=1"`
}

// PublishAck is returned to the provider on success.
type PublishAck struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// BecknPublishRequest is the Beckn-compliant payload sent from the BPP to the CDS.
type BecknPublishRequest struct {
	Context BecknContext        `json:"context"`
	Message BecknPublishMessage `json:"message"`
}

type BecknContext struct {
	Version       string `json:"version"`
	Action        string `json:"action"`
	Timestamp     string `json:"timestamp"`
	TransactionID string `json:"transactionId"`
	MessageID     string `json:"messageId"`
	BppID         string `json:"bppId"`
	BppURI        string `json:"bppUri"`
	NetworkID     string `json:"networkId"`
}

type BecknPublishMessage struct {
	Catalogs []Catalog `json:"catalogs"`
}

// ---------------------------------------------------------------------------
// Shared domain types (used in both provider request and CDS forward payload)
// ---------------------------------------------------------------------------

type Catalog struct {
	ID                string             `json:"id"`
	Descriptor        Descriptor         `json:"descriptor"`
	Provider          Provider           `json:"provider"`
	Resources         []Resource         `json:"resources,omitempty"`
	Offers            []Offer            `json:"offers,omitempty"`
	PublishDirectives *PublishDirectives `json:"publishDirectives,omitempty"`
}

type Descriptor struct {
	Name      string      `json:"name"`
	Code      string      `json:"code,omitempty"`
	ShortDesc string      `json:"shortDesc,omitempty"`
	LongDesc  string      `json:"longDesc,omitempty"`
	MediaFile []MediaFile `json:"mediaFile,omitempty"`
}

type MediaFile struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType"`
	Label    string `json:"label,omitempty"`
}

type Provider struct {
	ID          string     `json:"id"`
	Descriptor  Descriptor `json:"descriptor"`
	AvailableAt []Location `json:"availableAt,omitempty"`
}

type Location struct {
	Geo     GeoJSON `json:"geo"`
	Address Address `json:"address"`
}

type GeoJSON struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type Address struct {
	StreetAddress   string `json:"streetAddress,omitempty"`
	AddressLocality string `json:"addressLocality,omitempty"`
	AddressRegion   string `json:"addressRegion,omitempty"`
	PostalCode      string `json:"postalCode,omitempty"`
	AddressCountry  string `json:"addressCountry,omitempty"`
}

type Resource struct {
	ID                 string          `json:"id"`
	Descriptor         Descriptor      `json:"descriptor"`
	ResourceAttributes json.RawMessage `json:"resourceAttributes,omitempty"`
}

type Offer struct {
	ID              string          `json:"id"`
	Descriptor      Descriptor      `json:"descriptor"`
	Provider        *Provider       `json:"provider,omitempty"`
	ResourceIDs     []string        `json:"resourceIds"`
	Considerations  []Consideration `json:"considerations,omitempty"`
	Validity        *TimePeriod     `json:"validity,omitempty"`
	OfferAttributes json.RawMessage `json:"offerAttributes,omitempty"`
}

type Consideration struct {
	ID                      string          `json:"id"`
	Status                  Status          `json:"status"`
	ConsiderationAttributes json.RawMessage `json:"considerationAttributes,omitempty"`
}

type Status struct {
	Name string `json:"name"`
	Code string `json:"code,omitempty"`
}

type TimePeriod struct {
	StartDate string `json:"startDate,omitempty"`
	EndDate   string `json:"endDate,omitempty"`
}

type PublishDirectives struct {
	CatalogType string `json:"catalogType,omitempty"`
}
