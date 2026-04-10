package clientsvc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
		respBody []byte
		callErr  error
	)

	// Deferred log captures the final error and full response body.
	defer func() {
		ms := int32(time.Since(start).Milliseconds())
		ack := dbsqlc.NullAckStatus{AckStatus: dbsqlc.AckStatusACK, Valid: true}
		if callErr != nil {
			ack = dbsqlc.NullAckStatus{AckStatus: dbsqlc.AckStatusNACK, Valid: true}
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
			s.lh.WithModule("client_svc").Warn().Error(err).Log("failed to write discover message log")
		}
	}()

	s.lh.WithModule("client_svc").Log(fmt.Sprintf("sending discover to %s", s.cfg.CDSDiscoverURL))

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

// Select initiates a Select action to the BPP.
func (s *ClientService) Select(ctx context.Context, req *ClientSelectRequest) (string, error) {
	txnID := uuid.New()
	msgID := uuid.New()

	becknReq := BecknRequest{
		Context: s.newContext("select", txnID.String(), msgID.String(), req.BppID, req.BppURI),
		Message: json.RawMessage(fmt.Sprintf(`{"order":{"items":%s}}`, string(req.Items))),
	}

	if err := s.sendToBPP(ctx, req.BppURI, "select", txnID, msgID, req.NetworkID, becknReq); err != nil {
		return "", err
	}

	// 3. Update transaction status
	q := dbsqlc.New(s.pool)
	if err := q.UpsertTransaction(ctx, dbsqlc.UpsertTransactionParams{
		TransactionID: txnID,
		BapID:         s.cfg.BapID,
		NetworkID:     &req.NetworkID,
		BppID:         &req.BppID,
		BppUri:        &req.BppURI,
		Status:        dbsqlc.TransactionStatusSELECTSENT,
	}); err != nil {
		s.lh.WithModule("client_svc").Warn().Error(err).Log("upsert transaction failed")
	}

	return txnID.String(), nil
}

// Init initiates an Init action to the BPP.
func (s *ClientService) Init(ctx context.Context, req *ClientInitRequest) error {
	txnID, _ := uuid.Parse(req.TransactionID)
	msgID := uuid.New()

	becknReq := BecknRequest{
		Context: s.newContext("init", req.TransactionID, msgID.String(), req.BppID, req.BppURI),
		Message: json.RawMessage(fmt.Sprintf(`{"order":{"billing":%s,"fulfillments":%s}}`, string(req.Billing), string(req.Fulfillments))),
	}

	if err := s.sendToBPP(ctx, req.BppURI, "init", txnID, msgID, req.NetworkID, becknReq); err != nil {
		return err
	}

	q := dbsqlc.New(s.pool)
	return q.UpsertTransaction(ctx, dbsqlc.UpsertTransactionParams{
		TransactionID: txnID,
		BapID:         s.cfg.BapID,
		NetworkID:     &req.NetworkID,
		BppID:         &req.BppID,
		BppUri:        &req.BppURI,
		Status:        dbsqlc.TransactionStatusINITSENT,
	})
}

// Confirm initiates a Confirm action to the BPP.
func (s *ClientService) Confirm(ctx context.Context, req *ClientConfirmRequest) error {
	txnID, _ := uuid.Parse(req.TransactionID)
	msgID := uuid.New()

	becknReq := BecknRequest{
		Context: s.newContext("confirm", req.TransactionID, msgID.String(), req.BppID, req.BppURI),
		Message: json.RawMessage(`{"order":{}}`), // Empty order for confirm (snapshot-based)
	}

	if err := s.sendToBPP(ctx, req.BppURI, "confirm", txnID, msgID, req.NetworkID, becknReq); err != nil {
		return err
	}

	q := dbsqlc.New(s.pool)
	return q.UpsertTransaction(ctx, dbsqlc.UpsertTransactionParams{
		TransactionID: txnID,
		BapID:         s.cfg.BapID,
		NetworkID:     &req.NetworkID,
		BppID:         &req.BppID,
		BppUri:        &req.BppURI,
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
	}
}

// sendToBPP transmits a Beckn request to a BPP and logs both the outbound request
// and the BPP ACK response to beckn_message_log.
func (s *ClientService) sendToBPP(ctx context.Context, bppURI, action string, txnID, msgID uuid.UUID, networkID string, req BecknRequest) error {
	targetURL := fmt.Sprintf("%s/%s", bppURI, action)
	reqJSON, _ := json.Marshal(req)

	s.lh.WithModule("client_svc").Log(fmt.Sprintf("sending %s to %s", action, targetURL))

	start := time.Now()

	var (
		respBody []byte
		callErr  error
	)

	defer func() {
		ms := int32(time.Since(start).Milliseconds())
		ack := dbsqlc.NullAckStatus{AckStatus: dbsqlc.AckStatusACK, Valid: true}
		if callErr != nil {
			ack = dbsqlc.NullAckStatus{AckStatus: dbsqlc.AckStatusNACK, Valid: true}
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
			s.lh.WithModule("client_svc").Warn().Error(err).Log("failed to log outbound message")
		}
	}()

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

	respBody, _ = io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		callErr = fmt.Errorf("BPP returned status %d: %s", resp.StatusCode, string(respBody))
		return callErr
	}

	return nil
}

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
