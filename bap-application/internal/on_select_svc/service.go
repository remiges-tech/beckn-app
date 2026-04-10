package onselectsvc

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

// OnSelectService persists the quoted contract snapshot received from the BPP's on_select.
type OnSelectService struct {
	pool *pgxpool.Pool
	cfg  *config.Config
	lh   *logharbour.Logger
}

func NewOnSelectService(pool *pgxpool.Pool, cfg *config.Config, lh *logharbour.Logger) *OnSelectService {
	return &OnSelectService{pool: pool, cfg: cfg, lh: lh}
}

// Process persists the on_select snapshot and writes to beckn_message_log.
func (s *OnSelectService) Process(ctx context.Context, req *OnSelectRequest) {
	start := time.Now()
	reqJSON, _ := json.Marshal(req)
	txID, _ := uuid.Parse(req.Context.TransactionID)
	msgID, _ := uuid.Parse(req.Context.MessageID)
	q := dbsqlc.New(s.pool)

	// Upsert transaction row (creates it if it doesn't exist from a prior step).
	if err := q.UpsertTransaction(ctx, dbsqlc.UpsertTransactionParams{
		TransactionID: txID,
		BapID:         s.cfg.BapID,
		NetworkID:     strPtr(req.Context.NetworkID),
		BppID:         strPtr(req.Context.BppID),
		BppUri:        strPtr(req.Context.BppURI),
		Status:        dbsqlc.TransactionStatusQUOTERECEIVED,
	}); err != nil {
		s.lh.WithModule("on_select_svc").Warn().Error(err).Log("upsert transaction failed")
	}

	// Store the contract snapshot.
	contractJSON := req.Message.Contract
	if len(contractJSON) == 0 {
		contractJSON = json.RawMessage("null")
	}
	if err := q.UpsertContractSnapshot(ctx, dbsqlc.UpsertContractSnapshotParams{
		TransactionID: txID,
		Action:        dbsqlc.BecknActionOnSelect,
		ContractID:    nil, // BPP assigns contract id during on_init, not on_select
		Contract:      contractJSON,
	}); err != nil {
		s.lh.WithModule("on_select_svc").Warn().Error(err).Log("upsert contract snapshot failed")
	}

	// Audit log.
	ms := int32(time.Since(start).Milliseconds())
	if err := q.InsertMessageLog(ctx, dbsqlc.InsertMessageLogParams{
		TransactionID:        pgtype.UUID{Bytes: txID, Valid: true},
		MessageID:            msgID,
		Action:               dbsqlc.BecknActionOnSelect,
		Direction:            dbsqlc.MessageDirectionINBOUND,
		Url:                  strPtr("/webhook/on_select"),
		BapID:                strPtr(req.Context.BapID),
		BapUri:               strPtr(req.Context.BapURI),
		BppID:                strPtr(req.Context.BppID),
		BppUri:               strPtr(req.Context.BppURI),
		NetworkID:            strPtr(req.Context.NetworkID),
		AckStatus:            dbsqlc.NullAckStatus{AckStatus: dbsqlc.AckStatusACK, Valid: true},
		RequestPayload:       reqJSON,
		ProcessingDurationMs: &ms,
	}); err != nil {
		s.lh.WithModule("on_select_svc").Warn().Error(err).Log(fmt.Sprintf("message log insert failed: %v", err))
	}

	s.lh.WithModule("on_select_svc").Log("on_select processed for txn " + req.Context.TransactionID)
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
