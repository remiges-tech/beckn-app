package oninitsvc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/remiges-tech/logharbour/logharbour"

	dbsqlc "github.com/ion/winroom/bap/db/sqlc"
	"github.com/ion/winroom/bap/internal/config"
)

// OnInitService persists the on_init contract snapshot (which carries the BPP-assigned contract id).
type OnInitService struct {
	pool *pgxpool.Pool
	cfg  *config.Config
	lh   *logharbour.Logger
}

func NewOnInitService(pool *pgxpool.Pool, cfg *config.Config, lh *logharbour.Logger) *OnInitService {
	return &OnInitService{pool: pool, cfg: cfg, lh: lh}
}

// Process persists the on_init snapshot and writes to beckn_message_log.
func (s *OnInitService) Process(ctx context.Context, req *OnInitRequest) {
	start := time.Now()
	reqJSON, _ := json.Marshal(req)
	txID, _ := uuid.Parse(req.Context.TransactionID)
	msgID, _ := uuid.Parse(req.Context.MessageID)
	q := dbsqlc.New(s.pool)

	// Extract the contract id from the contract JSON (BPP assigned it during init).
	contractID := extractContractID(req.Message.Contract)

	// Upsert transaction with INIT_RECEIVED status.
	if err := q.UpsertTransaction(ctx, dbsqlc.UpsertTransactionParams{
		TransactionID: txID,
		BapID:         s.cfg.BapID,
		NetworkID:     strPtr(req.Context.NetworkID),
		BppID:         strPtr(req.Context.BppID),
		BppUri:        strPtr(req.Context.BppURI),
		Status:        dbsqlc.TransactionStatusINITRECEIVED,
	}); err != nil {
		s.lh.WithModule("on_init_svc").Warn().Error(err).Log("upsert transaction failed")
	}

	// Store contract snapshot under 'on_init', keeping the BPP contract id.
	contractJSON := req.Message.Contract
	if len(contractJSON) == 0 {
		contractJSON = json.RawMessage("null")
	}
	if err := q.UpsertContractSnapshot(ctx, dbsqlc.UpsertContractSnapshotParams{
		TransactionID: txID,
		Action:        dbsqlc.BecknActionOnInit,
		ContractID:    contractID,
		Contract:      contractJSON,
	}); err != nil {
		s.lh.WithModule("on_init_svc").Warn().Error(err).Log("upsert contract snapshot failed")
	}

	// Audit log.
	ms := int32(time.Since(start).Milliseconds())
	if err := q.InsertMessageLog(ctx, dbsqlc.InsertMessageLogParams{
		TransactionID:        pgtype.UUID{Bytes: txID, Valid: true},
		MessageID:            msgID,
		Action:               dbsqlc.BecknActionOnInit,
		Direction:            dbsqlc.MessageDirectionINBOUND,
		Url:                  strPtr("/webhook/on_init"),
		BapID:                strPtr(req.Context.BapID),
		BapUri:               strPtr(req.Context.BapURI),
		BppID:                strPtr(req.Context.BppID),
		BppUri:               strPtr(req.Context.BppURI),
		NetworkID:            strPtr(req.Context.NetworkID),
		AckStatus:            dbsqlc.NullAckStatus{AckStatus: dbsqlc.AckStatusACK, Valid: true},
		RequestPayload:       reqJSON,
		ProcessingDurationMs: &ms,
	}); err != nil {
		s.lh.WithModule("on_init_svc").Warn().Error(err).Log(fmt.Sprintf("message log insert failed: %v", err))
	}

	s.lh.WithModule("on_init_svc").Log("on_init processed for txn " + req.Context.TransactionID + " contractId=" + safeStr(contractID))
}

// extractContractID pulls the "id" field from the contract JSON blob.
func extractContractID(raw json.RawMessage) *string {
	if len(raw) == 0 {
		return nil
	}
	var c struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &c); err != nil || c.ID == "" {
		return nil
	}
	return &c.ID
}

func safeStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
