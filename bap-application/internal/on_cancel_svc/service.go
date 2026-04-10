package oncancelsvc

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

type Service struct {
	pool *pgxpool.Pool
	cfg  *config.Config
	lh   *logharbour.Logger
}

func NewService(pool *pgxpool.Pool, cfg *config.Config, lh *logharbour.Logger) *Service {
	return &Service{pool: pool, cfg: cfg, lh: lh}
}

// Process transitions the transaction to CANCELLED and stores the snapshot.
func (s *Service) Process(ctx context.Context, req *OnCancelRequest) {
	start := time.Now()
	reqJSON, _ := json.Marshal(req)
	txID, _ := uuid.Parse(req.Context.TransactionID)
	msgID, _ := uuid.Parse(req.Context.MessageID)
	q := dbsqlc.New(s.pool)

	// Advance transaction status to CANCELLED.
	if err := q.UpsertTransaction(ctx, dbsqlc.UpsertTransactionParams{
		TransactionID: txID,
		BapID:         s.cfg.BapID,
		NetworkID:     strPtr(req.Context.NetworkID),
		BppID:         strPtr(req.Context.BppID),
		BppUri:        strPtr(req.Context.BppURI),
		Status:        dbsqlc.TransactionStatusCANCELLED,
	}); err != nil {
		s.lh.WithModule("on_cancel_svc").Warn().Error(err).Log("upsert transaction failed")
	}

	contractJSON := req.Message.Contract
	if len(contractJSON) == 0 {
		contractJSON = json.RawMessage("null")
	}
	contractID := extractContractID(contractJSON)
	if err := q.UpsertContractSnapshot(ctx, dbsqlc.UpsertContractSnapshotParams{
		TransactionID: txID,
		Action:        dbsqlc.BecknActionOnCancel,
		ContractID:    contractID,
		Contract:      contractJSON,
	}); err != nil {
		s.lh.WithModule("on_cancel_svc").Warn().Error(err).Log("upsert contract snapshot failed")
	}

	ms := int32(time.Since(start).Milliseconds())
	if err := q.InsertMessageLog(ctx, dbsqlc.InsertMessageLogParams{
		TransactionID:        pgtype.UUID{Bytes: txID, Valid: true},
		MessageID:            msgID,
		Action:               dbsqlc.BecknActionOnCancel,
		Direction:            dbsqlc.MessageDirectionINBOUND,
		Url:                  strPtr("/api/webhook/on_cancel"),
		BapID:                strPtr(req.Context.BapID),
		BapUri:               strPtr(req.Context.BapURI),
		BppID:                strPtr(req.Context.BppID),
		BppUri:               strPtr(req.Context.BppURI),
		NetworkID:            strPtr(req.Context.NetworkID),
		AckStatus:            dbsqlc.NullAckStatus{AckStatus: dbsqlc.AckStatusACK, Valid: true},
		RequestPayload:       reqJSON,
		ProcessingDurationMs: &ms,
	}); err != nil {
		s.lh.WithModule("on_cancel_svc").Warn().Error(err).Log(fmt.Sprintf("message log insert failed: %v", err))
	}

	s.lh.WithModule("on_cancel_svc").LogActivity(
		fmt.Sprintf("← IN   ON_CANCEL | txn=%.8s… | bpp=%s | dur=%dms", req.Context.TransactionID, req.Context.BppID, time.Since(start).Milliseconds()),
		json.RawMessage(reqJSON),
	)
}

func extractContractID(raw json.RawMessage) *string {
	if len(raw) == 0 {
		return nil
	}
	var c struct{ ID string `json:"id"` }
	if err := json.Unmarshal(raw, &c); err != nil || c.ID == "" {
		return nil
	}
	return &c.ID
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
