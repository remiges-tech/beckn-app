package clientsvc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/remiges-tech/logharbour/logharbour"

	dbsqlc "github.com/ion/winroom/bap/db/sqlc"
	"github.com/ion/winroom/bap/internal/config"
)

type ClientService struct {
	pool   *pgxpool.Pool
	cfg    *config.Config
	lh     *logharbour.Logger
	client *http.Client
}

func NewClientService(pool *pgxpool.Pool, cfg *config.Config, lh *logharbour.Logger) *ClientService {
	return &ClientService{
		pool: pool,
		cfg:  cfg,
		lh:   lh,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Discover sends a Beckn v2 `discover` action to the CDS and returns the raw JSON response.
// Logs an OUTBOUND row to beckn_message_log capturing both the request and the CDS response.
func (s *ClientService) Discover(ctx context.Context, req *ClientDiscoverRequest) (json.RawMessage, error) {
	if s.cfg.CDSDiscoverURL == "" {
		return json.RawMessage(`{"message":{"catalogs":[]}}`), nil
	}

	start := time.Now()
	msgID := uuid.New()
	txnID := uuid.New()

	becknReq := map[string]interface{}{
		"context": map[string]interface{}{
			"version":       "2.0.0",
			"action":        "discover",
			"timestamp":     start.UTC().Format(time.RFC3339),
			"messageId":     msgID.String(),
			"transactionId": txnID.String(),
			"bapId":         s.cfg.BapID,
			"bapUri":        s.cfg.BapURI,
			"networkId":     s.cfg.NetworkID,
		},
		"message": map[string]interface{}{
			"intent": map[string]interface{}{
				"textSearch": req.TextSearch,
			},
		},
	}

	reqJSON, _ := json.Marshal(becknReq)

	var (
		respBody   []byte
		callErr    error
		httpStatus int
	)

	// Deferred block: DB log + logharbour ACK/ERR entry.
	defer func() {
		dur := time.Since(start)
		ms := int32(dur.Milliseconds())
		ack := dbsqlc.NullAckStatus{AckStatus: dbsqlc.AckStatusACK, Valid: true}
		if callErr != nil {
			ack = dbsqlc.NullAckStatus{AckStatus: dbsqlc.AckStatusNACK, Valid: true}
			s.logBecknErr("discover", txnID.String(), s.cfg.CDSDiscoverURL, callErr, dur)
		} else {
			s.logBecknAck("discover", txnID.String(), httpStatus, respBody, dur)
		}
		q := dbsqlc.New(s.pool)
		if err := q.InsertMessageLog(ctx, dbsqlc.InsertMessageLogParams{
			TransactionID:        pgtype.UUID{Bytes: txnID, Valid: true},
			MessageID:            msgID,
			Action:               dbsqlc.BecknActionDiscover,
			Direction:            dbsqlc.MessageDirectionOUTBOUND,
			Url:                  strPtr(s.cfg.CDSDiscoverURL),
			BapID:                strPtr(s.cfg.BapID),
			BapUri:               strPtr(s.cfg.BapURI),
			NetworkID:            strPtr(s.cfg.NetworkID),
			AckStatus:            ack,
			RequestPayload:       reqJSON,
			ResponsePayload:      json.RawMessage(respBody),
			ErrorMessage:         errStrPtr(callErr),
			ProcessingDurationMs: &ms,
		}); err != nil {
			s.lh.WithModule("beckn_client").Warn().Error(err).Log("failed to write discover message log")
		}
	}()

	s.logBecknOut("discover", http.MethodGet, s.cfg.CDSDiscoverURL, txnID.String(), msgID.String(), reqJSON)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, s.cfg.CDSDiscoverURL, bytes.NewBuffer(reqJSON))
	if err != nil {
		callErr = fmt.Errorf("build discover request: %w", err)
		return nil, callErr
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		callErr = fmt.Errorf("discover HTTP call failed: %w", err)
		return nil, callErr
	}
	defer resp.Body.Close() //nolint:errcheck

	httpStatus = resp.StatusCode

	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		callErr = fmt.Errorf("read discover response: %w", err)
		return nil, callErr
	}

	if resp.StatusCode != http.StatusOK {
		callErr = fmt.Errorf("CDS returned status %d: %s", resp.StatusCode, string(respBody))
		return nil, callErr
	}

	return respBody, nil
}

// Select initiates a Select action to the ONIX BAP adapter.
// Builds a Beckn v2 message.contract from the frontend's SelectItem list.
func (s *ClientService) Select(ctx context.Context, req *ClientSelectRequest) (string, error) {
	txnID := uuid.New()
	msgID := uuid.New()

	msgJSON, err := buildSelectMessage(req.Items, s.cfg.BapID)
	if err != nil {
		return "", fmt.Errorf("build select message: %w", err)
	}

	becknReq := BecknRequest{
		Context: s.newContext("select", txnID.String(), msgID.String(), s.cfg.BppID, s.cfg.BppURI),
		Message: json.RawMessage(msgJSON),
	}

	if err := s.sendToBPP(ctx, s.cfg.AdapterURL, "select", txnID, msgID, s.cfg.NetworkID, becknReq); err != nil {
		return "", err
	}

	q := dbsqlc.New(s.pool)
	if err := q.UpsertTransaction(ctx, dbsqlc.UpsertTransactionParams{
		TransactionID: txnID,
		BapID:         s.cfg.BapID,
		NetworkID:     &s.cfg.NetworkID,
		BppID:         &s.cfg.BppID,
		BppUri:        &s.cfg.BppURI,
		Status:        dbsqlc.TransactionStatusSELECTSENT,
	}); err != nil {
		s.lh.WithModule("client_svc").Warn().Error(err).Log("upsert transaction failed")
	}

	return txnID.String(), nil
}

// buildSelectMessage constructs the Beckn v2 message.contract JSON for a select call.
func buildSelectMessage(items []SelectItem, bapID string) ([]byte, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("at least one item required")
	}

	// Use first item's provider for the provider participant (single-provider per select).
	providerID := items[0].ProviderID
	if providerID == "" {
		providerID = "provider-default"
	}
	providerName := items[0].ProviderName
	buyerID := "buyer-" + bapID

	// Build commitments, one per item.
	commitments := make([]map[string]interface{}, 0, len(items))
	commitmentIDs := make([]string, 0, len(items))
	for i, item := range items {
		cid := fmt.Sprintf("commitment-%s", item.ResourceID)
		commitmentIDs = append(commitmentIDs, cid)

		qty := item.Quantity
		if qty <= 0 {
			qty = 1
		}

		// Build a sanitized offerAttributes containing only @context, @type and policies.
		// The raw offerAttributes from the CDS discover response may contain fields
		// (e.g. price.applicableQuantity.minValue) that the adapter's strict schema rejects.
		offerAttr := sanitizeOfferAttributes(item.OfferAttributes)

		offerName := item.OfferName
		if offerName == "" {
			offerName = item.ResourceID
		}

		commitments = append(commitments, map[string]interface{}{
			"id":     cid,
			"status": map[string]interface{}{"descriptor": map[string]interface{}{"code": "DRAFT"}},
			"resources": []map[string]interface{}{
				{"id": item.ResourceID},
			},
			"offer": map[string]interface{}{
				"id":          item.OfferID,
				"resourceIds": []string{item.ResourceID},
				"descriptor":  map[string]interface{}{"name": offerName},
				"provider": map[string]interface{}{
					"id":         providerID,
					"descriptor": map[string]interface{}{"name": providerName},
				},
				"offerAttributes": offerAttr,
			},
			"commitmentAttributes": map[string]interface{}{
				"@context":   "https://raw.githubusercontent.com/beckn/local-retail/refs/heads/main/schema/RetailCommitment/v2.1/context.jsonld",
				"@type":      "RetailCommitment",
				"lineId":     fmt.Sprintf("LINE-%03d", i+1),
				"resourceId": item.ResourceID,
				"quantity":   map[string]interface{}{"unitCode": "EA", "unitQuantity": qty},
			},
		})
	}

	contract := map[string]interface{}{
		"status": map[string]interface{}{"code": "DRAFT"},
		"participants": []map[string]interface{}{
			{
				"id":          providerID,
				"descriptor":  map[string]interface{}{"name": providerName},
				"participantAttributes": map[string]interface{}{
					"@context": "https://raw.githubusercontent.com/beckn/schemas/main/schema/Provider/v2.1/context.jsonld",
					"@type":    "Provider",
					"id":       providerID,
					"descriptor": map[string]interface{}{
						"name": providerName,
					},
				},
			},
			{
				"id":         buyerID,
				"descriptor": map[string]interface{}{"name": "Buyer"},
				"participantAttributes": map[string]interface{}{
					"@context": "https://raw.githubusercontent.com/beckn/schemas/main/schema/Consumer/v2.0/context.jsonld",
					"@type":    "Consumer",
					"role":     "buyer",
					"person":   map[string]interface{}{"id": buyerID, "name": "Buyer"},
				},
			},
		},
		"commitments": commitments,
		"performance": []map[string]interface{}{
			{
				"id":            "f1",
				"status":        map[string]interface{}{"code": "PENDING"},
				"commitmentIds": commitmentIDs,
				"performanceAttributes": map[string]interface{}{
					"@context": "https://raw.githubusercontent.com/beckn/local-retail/refs/heads/main/schema/RetailPerformance/v2.1/context.jsonld",
					"@type":    "rcpa:RetailPerformance",
					"supportedPerformanceModes": []string{"DELIVERY"},
					"deliveryDetails": map[string]interface{}{
						"address": map[string]interface{}{
							"streetAddress":   "TBD",
							"addressLocality": "TBD",
							"addressRegion":   "TBD",
							"postalCode":      "000000",
							"addressCountry":  "IN",
						},
						"contact": map[string]interface{}{
							"name":  "Buyer",
							"phone": "+910000000000",
						},
					},
					"operatingHours": []map[string]interface{}{
						{
							"daysOfWeek": []int{1, 2, 3, 4, 5, 6},
							"timeRange":  map[string]string{"start": "09:00", "end": "21:00"},
						},
					},
					"installationScheduling": map[string]interface{}{"available": false},
				},
			},
		},
	}

	return json.Marshal(map[string]interface{}{"contract": contract})
}

// sanitizeOfferAttributes extracts only @context, @type, and policies from the raw
// offerAttributes discovered via the CDS. The adapter's strict JSON-LD schema rejects
// fields like price.applicableQuantity.minValue that the CDS uses but the adapter does not.
func sanitizeOfferAttributes(raw json.RawMessage) map[string]interface{} {
	const (
		offerCtx  = "https://raw.githubusercontent.com/beckn/local-retail/refs/heads/main/schema/RetailOffer/v2.1/context.jsonld"
		offerType = "RetailOffer"
	)

	out := map[string]interface{}{
		"@context": offerCtx,
		"@type":    offerType,
	}

	if len(raw) <= 2 {
		return out
	}

	var decoded map[string]json.RawMessage
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return out
	}

	// Preserve @context / @type from the discover response if present.
	if v, ok := decoded["@context"]; ok {
		var s string
		if json.Unmarshal(v, &s) == nil && s != "" {
			out["@context"] = s
		}
	}
	if v, ok := decoded["@type"]; ok {
		var s string
		if json.Unmarshal(v, &s) == nil && s != "" {
			out["@type"] = s
		}
	}

	// Carry policies through — returns and cancellation are safe per adapter schema.
	if v, ok := decoded["policies"]; ok {
		var policies interface{}
		if json.Unmarshal(v, &policies) == nil {
			out["policies"] = policies
		}
	}

	return out
}

// Init initiates an Init action to the ONIX BAP adapter.
func (s *ClientService) Init(ctx context.Context, req *ClientInitRequest) error {
	txnID, _ := uuid.Parse(req.TransactionID)
	msgID := uuid.New()

	becknReq := BecknRequest{
		Context: s.newContext("init", req.TransactionID, msgID.String(), s.cfg.BppID, s.cfg.BppURI),
		Message: json.RawMessage(fmt.Sprintf(`{"order":{"billing":%s,"fulfillments":%s}}`, string(req.Billing), string(req.Fulfillments))),
	}

	if err := s.sendToBPP(ctx, s.cfg.AdapterURL, "init", txnID, msgID, s.cfg.NetworkID, becknReq); err != nil {
		return err
	}

	q := dbsqlc.New(s.pool)
	return q.UpsertTransaction(ctx, dbsqlc.UpsertTransactionParams{
		TransactionID: txnID,
		BapID:         s.cfg.BapID,
		NetworkID:     &s.cfg.NetworkID,
		BppID:         &s.cfg.BppID,
		BppUri:        &s.cfg.BppURI,
		Status:        dbsqlc.TransactionStatusINITSENT,
	})
}

// Confirm initiates a Confirm action to the ONIX BAP adapter.
func (s *ClientService) Confirm(ctx context.Context, req *ClientConfirmRequest) error {
	txnID, _ := uuid.Parse(req.TransactionID)
	msgID := uuid.New()

	becknReq := BecknRequest{
		Context: s.newContext("confirm", req.TransactionID, msgID.String(), s.cfg.BppID, s.cfg.BppURI),
		Message: json.RawMessage(`{"order":{}}`),
	}

	if err := s.sendToBPP(ctx, s.cfg.AdapterURL, "confirm", txnID, msgID, s.cfg.NetworkID, becknReq); err != nil {
		return err
	}

	q := dbsqlc.New(s.pool)
	return q.UpsertTransaction(ctx, dbsqlc.UpsertTransactionParams{
		TransactionID: txnID,
		BapID:         s.cfg.BapID,
		NetworkID:     &s.cfg.NetworkID,
		BppID:         &s.cfg.BppID,
		BppUri:        &s.cfg.BppURI,
		Status:        dbsqlc.TransactionStatusCONFIRMSENT,
	})
}

// GetStatus returns the latest state for a transaction.
func (s *ClientService) GetStatus(ctx context.Context, txnIDStr string) (*ClientStatusResponse, error) {
	txnID, err := uuid.Parse(txnIDStr)
	if err != nil {
		return nil, err
	}

	q := dbsqlc.New(s.pool)
	txn, err := q.GetTransaction(ctx, txnID)
	if err != nil {
		return nil, err
	}

	// Get latest snapshot if available
	snapshot, _ := q.GetLatestContractSnapshot(ctx, txnID)

	resp := &ClientStatusResponse{
		TransactionID: txnIDStr,
		Status:        string(txn.Status),
		UpdatedAt:     txn.UpdatedAt.Time,
	}

	if snapshot.ID != [16]byte{} {
		resp.Contract = snapshot.Contract
	}

	return resp, nil
}

// GetCatalog returns a mock catalog (for now) or triggers a search.
func (s *ClientService) GetCatalog(ctx context.Context) (json.RawMessage, error) {
	// Simple mock food catalog for the Instamart-style UI.
	return json.RawMessage(`[
		{"id":"p1","name":"Organic Avocados","price":120,"category":"Produce","image":"https://images.unsplash.com/photo-1523049673857-eb18f1d7b578?w=400&auto=format&fit=crop&q=60"},
		{"id":"p2","name":"Oat Milk (1L)","price":250,"category":"Dairy","image":"https://images.unsplash.com/photo-1634140785135-2465e9754ce4?w=400&auto=format&fit=crop&q=60"},
		{"id":"p3","name":"Greek Yogurt","price":80,"category":"Dairy","image":"https://images.unsplash.com/photo-1488477181946-6428a0291777?w=400&auto=format&fit=crop&q=60"},
		{"id":"p4","name":"Whole Wheat Bread","price":45,"category":"Bakery","image":"https://images.unsplash.com/photo-1549931319-a545dcf3bc73?w=400&auto=format&fit=crop&q=60"},
		{"id":"p5","name":"Dark Chocolate (70%)","price":150,"category":"Snacks","image":"https://images.unsplash.com/photo-1511381939415-e44015466834?w=400&auto=format&fit=crop&q=60"}
	]`), nil
}

// --- Helpers ---

func (s *ClientService) newContext(action, txnID, msgID, bppID, bppURI string) BecknContext {
	return BecknContext{
		Version:       "2.0.0",
		Action:        action,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		MessageID:     msgID,
		TransactionID: txnID,
		BapID:         s.cfg.BapID,
		BapURI:        s.cfg.BapURI,
		BppID:         bppID,
		BppURI:        bppURI,
		NetworkID:     s.cfg.NetworkID,
		TTL:           "PT30S",
	}
}

// sendToBPP transmits a Beckn request to the ONIX adapter and logs both the
// outbound request and the adapter ACK response to beckn_message_log and logharbour.
func (s *ClientService) sendToBPP(ctx context.Context, bppURI, action string, txnID, msgID uuid.UUID, networkID string, req BecknRequest) error {
	targetURL := fmt.Sprintf("%s/%s", bppURI, action)
	reqJSON, _ := json.Marshal(req)

	start := time.Now()

	var (
		respBody   []byte
		callErr    error
		httpStatus int
	)

	// Deferred block: DB log + logharbour ACK/ERR entry.
	defer func() {
		dur := time.Since(start)
		ms := int32(dur.Milliseconds())
		ack := dbsqlc.NullAckStatus{AckStatus: dbsqlc.AckStatusACK, Valid: true}
		if callErr != nil {
			ack = dbsqlc.NullAckStatus{AckStatus: dbsqlc.AckStatusNACK, Valid: true}
			s.logBecknErr(action, txnID.String(), targetURL, callErr, dur)
		} else {
			s.logBecknAck(action, txnID.String(), httpStatus, respBody, dur)
		}
		q := dbsqlc.New(s.pool)
		if err := q.InsertMessageLog(ctx, dbsqlc.InsertMessageLogParams{
			TransactionID:        pgtype.UUID{Bytes: txnID, Valid: true},
			MessageID:            msgID,
			Action:               dbsqlc.BecknAction(action),
			Direction:            dbsqlc.MessageDirectionOUTBOUND,
			Url:                  strPtr(targetURL),
			BapID:                strPtr(s.cfg.BapID),
			BapUri:               strPtr(s.cfg.BapURI),
			BppID:                &bppURI,
			BppUri:               &bppURI,
			NetworkID:            &networkID,
			AckStatus:            ack,
			RequestPayload:       reqJSON,
			ResponsePayload:      json.RawMessage(respBody),
			ErrorMessage:         errStrPtr(callErr),
			ProcessingDurationMs: &ms,
		}); err != nil {
			s.lh.WithModule("beckn_client").Warn().Error(err).Log("failed to log outbound message")
		}
	}()

	s.logBecknOut(action, http.MethodPost, targetURL, txnID.String(), msgID.String(), reqJSON)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewBuffer(reqJSON))
	if err != nil {
		callErr = err
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		callErr = err
		return err
	}
	defer resp.Body.Close() //nolint:errcheck

	httpStatus = resp.StatusCode
	respBody, _ = io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		callErr = fmt.Errorf("adapter returned status %d: %s", resp.StatusCode, string(respBody))
		return callErr
	}

	return nil
}

// --- Beckn call structured logging helpers ---

// becknOutActivity is the structured payload recorded before an outbound Beckn call.
type becknOutActivity struct {
	Action  string          `json:"action"`
	Method  string          `json:"method"`
	URL     string          `json:"url"`
	TxnID   string          `json:"txn_id"`
	MsgID   string          `json:"msg_id"`
	Request json.RawMessage `json:"request"`
}

// becknAckActivity is the structured payload recorded after a successful Beckn call.
type becknAckActivity struct {
	Action   string          `json:"action"`
	TxnID    string          `json:"txn_id"`
	HTTPCode int             `json:"http_code"`
	DurMs    int64           `json:"dur_ms"`
	Response json.RawMessage `json:"response"`
}

// becknErrActivity is the structured payload recorded when a Beckn call fails.
type becknErrActivity struct {
	Action string `json:"action"`
	TxnID  string `json:"txn_id"`
	URL    string `json:"url"`
	DurMs  int64  `json:"dur_ms"`
	Error  string `json:"error"`
}

// logBecknOut emits a logharbour Activity entry before an outbound Beckn call.
// The human-readable banner appears in "msg"; full request JSON is in data.activity_data.
func (s *ClientService) logBecknOut(action, method, url, txnID, msgID string, body []byte) {
	s.lh.WithModule("beckn_client").LogActivity(
		fmt.Sprintf("→ OUT  %-8s | txn=%.8s… | %s %s", strings.ToUpper(action), txnID, method, url),
		becknOutActivity{
			Action:  action,
			Method:  method,
			URL:     url,
			TxnID:   txnID,
			MsgID:   msgID,
			Request: json.RawMessage(body),
		},
	)
}

// logBecknAck emits a logharbour Activity entry after a successful Beckn call.
// The human-readable banner appears in "msg"; full response JSON is in data.activity_data.
func (s *ClientService) logBecknAck(action, txnID string, status int, body []byte, dur time.Duration) {
	s.lh.WithModule("beckn_client").LogActivity(
		fmt.Sprintf("← ACK  %-8s | txn=%.8s… | http=%d | dur=%dms", strings.ToUpper(action), txnID, status, dur.Milliseconds()),
		becknAckActivity{
			Action:   action,
			TxnID:    txnID,
			HTTPCode: status,
			DurMs:    dur.Milliseconds(),
			Response: json.RawMessage(body),
		},
	)
}

// logBecknErr emits a logharbour error Activity entry when a Beckn call fails.
func (s *ClientService) logBecknErr(action, txnID, url string, callErr error, dur time.Duration) {
	s.lh.WithModule("beckn_client").Err().Error(callErr).LogActivity(
		fmt.Sprintf("← ERR  %-8s | txn=%.8s… | dur=%dms | %s", strings.ToUpper(action), txnID, dur.Milliseconds(), callErr.Error()),
		becknErrActivity{
			Action: action,
			TxnID:  txnID,
			URL:    url,
			DurMs:  dur.Milliseconds(),
			Error:  callErr.Error(),
		},
	)
}

// --- Generic helpers ---

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func errStrPtr(err error) *string {
	if err == nil {
		return nil
	}
	s := err.Error()
	return &s
}
