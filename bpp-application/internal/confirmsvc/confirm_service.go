package confirmsvc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/remiges-tech/logharbour/logharbour"

	dbsqlc "github.com/ion/winroom/bpp/db/sqlc"
	"github.com/ion/winroom/bpp/internal/config"
)

// ConfirmService handles the full /beckn/confirm flow:
//  1. Transition the contract from DRAFT → ACTIVE in the database.
//  2. Build the on_confirm payload echoing back the confirmed contract.
//  3. POST the on_confirm callback to the BAP.
//  4. Write INBOUND and OUTBOUND audit rows to beckn_message_log.
type ConfirmService struct {
	pool       *pgxpool.Pool
	cfg        *config.Config
	lh         *logharbour.Logger
	httpClient *http.Client
}

func NewConfirmService(pool *pgxpool.Pool, cfg *config.Config, lh *logharbour.Logger) *ConfirmService {
	return &ConfirmService{
		pool:       pool,
		cfg:        cfg,
		lh:         lh,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// ProcessConfirm is called asynchronously after the BPP returns ACK to the BAP.
// It confirms the contract state, calls on_confirm on the BAP, and logs both messages.
func (s *ConfirmService) ProcessConfirm(ctx context.Context, req *ConfirmRequest) {
	start := time.Now()

	txID, _ := uuid.Parse(req.Context.TransactionID)
	inboundMsgID, _ := uuid.Parse(req.Context.MessageID)
	reqJSON, _ := json.Marshal(req)

	// Transition DRAFT → ACTIVE and retrieve the contract UUID.
	contractID, err := s.confirmContract(ctx, req)

	// Log INBOUND regardless of whether state transition succeeded.
	ackStatusInbound := dbsqlc.NullAckStatus{AckStatus: dbsqlc.AckStatusACK, Valid: true}
	var errMsg *string
	if err != nil {
		ackStatusInbound = dbsqlc.NullAckStatus{AckStatus: dbsqlc.AckStatusNACK, Valid: true}
		errMsg = errStrPtr(err)
		s.lh.WithModule("confirmsvc").Err().Error(err).Log("failed to confirm contract state")
	}

	s.writeMessageLog(ctx, dbsqlc.InsertMessageLogParams{
		TransactionID:        uuidToPgtype(txID),
		MessageID:            inboundMsgID,
		Action:               dbsqlc.BecknActionConfirm,
		Direction:            dbsqlc.MessageDirectionINBOUND,
		Url:                  strPtr("/webhook/confirm"),
		BapID:                strPtr(req.Context.BapID),
		BapUri:               strPtr(req.Context.BapURI),
		BppID:                strPtr(req.Context.BppID),
		BppUri:               strPtr(req.Context.BppURI),
		NetworkID:            strPtr(req.Context.NetworkID),
		AckStatus:            ackStatusInbound,
		RequestPayload:       reqJSON,
		ErrorMessage:         errMsg,
		ProcessingDurationMs: durationMs(start),
	})

	if err != nil {
		return
	}

	// Build the on_confirm payload — echo the contract back with ACTIVE status and the contract id.
	outboundMsgID := uuid.New()
	confirmedContract := req.Message.Contract
	confirmedContract.ID = contractID.String()
	confirmedContract.Status = &ContractStatus{Code: "ACTIVE"}

	onConfirmReq := OnConfirmRequest{
		Context: BecknContext{
			Version:       req.Context.Version,
			Action:        "on_confirm",
			Timestamp:     time.Now().UTC().Format(time.RFC3339),
			MessageID:     outboundMsgID.String(),
			TransactionID: req.Context.TransactionID,
			BapID:         req.Context.BapID,
			BapURI:        req.Context.BapURI,
			BppID:         req.Context.BppID,
			BppURI:        req.Context.BppURI,
			TTL:           req.Context.TTL,
			NetworkID:     req.Context.NetworkID,
		},
		Message: OnConfirmMessage{Contract: confirmedContract},
	}

	// POST on_confirm to BAP and log OUTBOUND.
	s.callOnConfirm(ctx, txID, outboundMsgID, req.Context.BapURI, onConfirmReq)
}

// confirmContract transitions the DRAFT contract to ACTIVE.
// Returns the UUID of the confirmed contract.
func (s *ConfirmService) confirmContract(ctx context.Context, req *ConfirmRequest) (uuid.UUID, error) {
	q := dbsqlc.New(s.pool)

	txID, err := uuid.Parse(req.Context.TransactionID)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid transactionId: %w", err)
	}

	contractID, err := q.ConfirmContract(ctx, dbsqlc.ConfirmContractParams{
		TransactionID: txID,
		BppID:         s.cfg.BppID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return uuid.UUID{}, fmt.Errorf("no DRAFT contract found for transactionId %q: confirm requires a prior init", req.Context.TransactionID)
		}
		return uuid.UUID{}, fmt.Errorf("confirm contract DB update: %w", err)
	}

	return contractID, nil
}

// callOnConfirm POSTs the on_confirm payload to the BAP and logs the OUTBOUND entry.
func (s *ConfirmService) callOnConfirm(
	ctx context.Context,
	txID uuid.UUID,
	msgID uuid.UUID,
	bapURI string,
	req OnConfirmRequest,
) {
	start := time.Now()
	onConfirmURL := bapURI + "/on_confirm"
	reqJSON, _ := json.Marshal(req)

	var (
		respBody []byte
		callErr  error
	)

	defer func() {
		ackStatus := dbsqlc.NullAckStatus{AckStatus: dbsqlc.AckStatusACK, Valid: true}
		if callErr != nil {
			ackStatus = dbsqlc.NullAckStatus{AckStatus: dbsqlc.AckStatusNACK, Valid: true}
		}
		s.writeMessageLog(ctx, dbsqlc.InsertMessageLogParams{
			TransactionID:        uuidToPgtype(txID),
			MessageID:            msgID,
			Action:               dbsqlc.BecknActionOnConfirm,
			Direction:            dbsqlc.MessageDirectionOUTBOUND,
			Url:                  strPtr(onConfirmURL),
			BapID:                strPtr(req.Context.BapID),
			BapUri:               strPtr(req.Context.BapURI),
			BppID:                strPtr(req.Context.BppID),
			BppUri:               strPtr(req.Context.BppURI),
			NetworkID:            strPtr(req.Context.NetworkID),
			AckStatus:            ackStatus,
			RequestPayload:       reqJSON,
			ResponsePayload:      rawMessageOrNull(respBody),
			ErrorMessage:         errStrPtr(callErr),
			ProcessingDurationMs: durationMs(start),
		})
	}()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, onConfirmURL, bytes.NewReader(reqJSON))
	if err != nil {
		callErr = fmt.Errorf("build on_confirm request: %w", err)
		s.lh.WithModule("confirmsvc").Err().Error(callErr).Log("on_confirm HTTP request build failed")
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		callErr = fmt.Errorf("on_confirm HTTP call to BAP failed: %w", err)
		s.lh.WithModule("confirmsvc").Err().Error(callErr).Log("on_confirm call failed")
		return
	}
	defer resp.Body.Close() //nolint:errcheck

	respBody, _ = io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		callErr = fmt.Errorf("BAP returned non-2xx status %d: %s", resp.StatusCode, string(respBody))
		s.lh.WithModule("confirmsvc").Warn().Error(callErr).Log("on_confirm BAP response error")
		return
	}

	s.lh.WithModule("confirmsvc").Log("on_confirm sent to BAP successfully")
}

// writeMessageLog inserts one audit row into beckn_message_log.
func (s *ConfirmService) writeMessageLog(ctx context.Context, params dbsqlc.InsertMessageLogParams) {
	q := dbsqlc.New(s.pool)
	if err := q.InsertMessageLog(ctx, params); err != nil {
		s.lh.WithModule("confirmsvc").Warn().Error(err).Log("failed to write beckn_message_log")
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func rawMessageOrNull(b []byte) json.RawMessage {
	if len(b) == 0 {
		return nil
	}
	return json.RawMessage(b)
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

func uuidToPgtype(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func durationMs(start time.Time) *int32 {
	ms := int32(time.Since(start).Milliseconds())
	return &ms
}
