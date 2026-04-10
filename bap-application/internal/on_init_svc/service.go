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

// becknInActivity is the structured payload logged when receiving an inbound Beckn callback.
type becknInActivity struct {
	Action  string          `json:"action"`
	TxnID   string          `json:"txn_id"`
	MsgID   string          `json:"msg_id"`
	BapID   string          `json:"bap_id"`
	BppID   string          `json:"bpp_id"`
	Payload json.RawMessage `json:"payload"`
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
		Url:                  strPtr("/api/webhook/on_init"),
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

	s.lh.WithModule("on_init_svc").LogActivity(
		fmt.Sprintf("← IN   ON_INIT   | txn=%.8s… | bpp=%s | contractId=%s | dur=%dms", req.Context.TransactionID, req.Context.BppID, safeStr(contractID), time.Since(start).Milliseconds()),
		becknInActivity{
			Action:  "on_init",
			TxnID:   req.Context.TransactionID,
			MsgID:   req.Context.MessageID,
			BapID:   req.Context.BapID,
			BppID:   req.Context.BppID,
			Payload: reqJSON,
		},
	)
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
