package onconfirmsvc

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

// OnConfirmService persists the confirmed (ACTIVE) contract snapshot from the BPP's on_confirm.
type OnConfirmService struct {
	pool *pgxpool.Pool
	cfg  *config.Config
	lh   *logharbour.Logger
}

func NewOnConfirmService(pool *pgxpool.Pool, cfg *config.Config, lh *logharbour.Logger) *OnConfirmService {
	return &OnConfirmService{pool: pool, cfg: cfg, lh: lh}
}

// Process persists the on_confirm snapshot and writes to beckn_message_log.
func (s *OnConfirmService) Process(ctx context.Context, req *OnConfirmRequest) {
	start := time.Now()
	reqJSON, _ := json.Marshal(req)
	txID, _ := uuid.Parse(req.Context.TransactionID)
	msgID, _ := uuid.Parse(req.Context.MessageID)
	q := dbsqlc.New(s.pool)

	contractID := extractContractID(req.Message.Contract)

	// Advance transaction to CONFIRMED.
	if err := q.UpsertTransaction(ctx, dbsqlc.UpsertTransactionParams{
		TransactionID: txID,
		BapID:         s.cfg.BapID,
		NetworkID:     strPtr(req.Context.NetworkID),
		BppID:         strPtr(req.Context.BppID),
		BppUri:        strPtr(req.Context.BppURI),
		Status:        dbsqlc.TransactionStatusCONFIRMED,
	}); err != nil {
		s.lh.WithModule("on_confirm_svc").Warn().Error(err).Log("upsert transaction failed")
	}

	// Store confirmed contract snapshot under 'on_confirm'.
	contractJSON := req.Message.Contract
	if len(contractJSON) == 0 {
		contractJSON = json.RawMessage("null")
	}
	if err := q.UpsertContractSnapshot(ctx, dbsqlc.UpsertContractSnapshotParams{
		TransactionID: txID,
		Action:        dbsqlc.BecknActionOnConfirm,
		ContractID:    contractID,
		Contract:      contractJSON,
	}); err != nil {
		s.lh.WithModule("on_confirm_svc").Warn().Error(err).Log("upsert contract snapshot failed")
	}

	// Audit log.
	ms := int32(time.Since(start).Milliseconds())
	if err := q.InsertMessageLog(ctx, dbsqlc.InsertMessageLogParams{
		TransactionID:        pgtype.UUID{Bytes: txID, Valid: true},
		MessageID:            msgID,
		Action:               dbsqlc.BecknActionOnConfirm,
		Direction:            dbsqlc.MessageDirectionINBOUND,
		Url:                  strPtr("/webhook/on_confirm"),
		BapID:                strPtr(req.Context.BapID),
		BapUri:               strPtr(req.Context.BapURI),
		BppID:                strPtr(req.Context.BppID),
		BppUri:               strPtr(req.Context.BppURI),
		NetworkID:            strPtr(req.Context.NetworkID),
		AckStatus:            dbsqlc.NullAckStatus{AckStatus: dbsqlc.AckStatusACK, Valid: true},
		RequestPayload:       reqJSON,
		ProcessingDurationMs: &ms,
	}); err != nil {
		s.lh.WithModule("on_confirm_svc").Warn().Error(err).Log(fmt.Sprintf("message log insert failed: %v", err))
	}

	s.lh.WithModule("on_confirm_svc").Log("on_confirm processed for txn " + req.Context.TransactionID + " contractId=" + safeStr(contractID))
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
