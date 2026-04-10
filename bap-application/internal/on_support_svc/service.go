package onsupportsvc

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

// Process stores the support ticket response in the message log.
func (s *Service) Process(ctx context.Context, req *OnSupportRequest) {
	start := time.Now()
	reqJSON, _ := json.Marshal(req)
	txID, _ := uuid.Parse(req.Context.TransactionID)
	msgID, _ := uuid.Parse(req.Context.MessageID)
	q := dbsqlc.New(s.pool)

	ms := int32(time.Since(start).Milliseconds())
	if err := q.InsertMessageLog(ctx, dbsqlc.InsertMessageLogParams{
		TransactionID:        pgtype.UUID{Bytes: txID, Valid: true},
		MessageID:            msgID,
		Action:               dbsqlc.BecknActionOnSupport,
		Direction:            dbsqlc.MessageDirectionINBOUND,
		Url:                  strPtr("/api/webhook/on_support"),
		BapID:                strPtr(req.Context.BapID),
		BapUri:               strPtr(req.Context.BapURI),
		BppID:                strPtr(req.Context.BppID),
		BppUri:               strPtr(req.Context.BppURI),
		NetworkID:            strPtr(req.Context.NetworkID),
		AckStatus:            dbsqlc.NullAckStatus{AckStatus: dbsqlc.AckStatusACK, Valid: true},
		RequestPayload:       reqJSON,
		ProcessingDurationMs: &ms,
	}); err != nil {
		s.lh.WithModule("on_support_svc").Warn().Error(err).Log(fmt.Sprintf("message log insert failed: %v", err))
	}

	s.lh.WithModule("on_support_svc").LogActivity(
		fmt.Sprintf("← IN   ON_SUPPORT| txn=%.8s… | bpp=%s | dur=%dms", req.Context.TransactionID, req.Context.BppID, time.Since(start).Milliseconds()),
		json.RawMessage(reqJSON),
	)
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
