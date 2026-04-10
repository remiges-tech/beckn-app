package cancelsvc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/remiges-tech/logharbour/logharbour"

	dbsqlc "github.com/ion/winroom/bpp/db/sqlc"
	"github.com/ion/winroom/bpp/internal/config"
)

type Service struct {
	pool       *pgxpool.Pool
	cfg        *config.Config
	lh         *logharbour.Logger
	httpClient *http.Client
}

func NewService(pool *pgxpool.Pool, cfg *config.Config, lh *logharbour.Logger) *Service {
	return &Service{pool: pool, cfg: cfg, lh: lh, httpClient: &http.Client{Timeout: 30 * time.Second}}
}

// Process cancels the ACTIVE contract, restores stock, and fires on_cancel to the BAP.
func (s *Service) Process(ctx context.Context, req *CancelRequest) {
	prettyLog(fmt.Sprintf("← IN   CANCEL   | txn=%.8s… | bap=%s", req.Context.TransactionID, req.Context.BapID), req)

	txID, err := uuid.Parse(req.Context.TransactionID)
	if err != nil {
		s.lh.WithModule("cancelsvc").Err().Error(err).Log("invalid transactionId")
		return
	}

	q := dbsqlc.New(s.pool)

	cancelled, err := q.CancelContract(ctx, dbsqlc.CancelContractParams{
		TransactionID: txID,
		BppID:         s.cfg.BppID,
	})
	if err != nil {
		s.lh.WithModule("cancelsvc").Err().Error(err).Log("cancel contract failed — not found or not ACTIVE")
		return
	}

	// Restore stock for all committed resources.
	s.restoreStock(ctx, q, cancelled.ID)

	contract := map[string]interface{}{
		"id":     cancelled.ID.String(),
		"status": map[string]string{"code": "CANCELLED"},
	}
	contractJSON, _ := json.Marshal(contract)

	outMsgID := uuid.New()
	onCancelReq := OnCancelRequest{
		Context: BecknContext{
			Version:       req.Context.Version,
			Action:        "on_cancel",
			Timestamp:     time.Now().UTC().Format(time.RFC3339),
			MessageID:     outMsgID.String(),
			TransactionID: req.Context.TransactionID,
			BapID:         req.Context.BapID,
			BapURI:        req.Context.BapURI,
			BppID:         req.Context.BppID,
			BppURI:        req.Context.BppURI,
			TTL:           req.Context.TTL,
			NetworkID:     req.Context.NetworkID,
		},
		Message: OnCancelMessage{Contract: contractJSON},
	}

	s.callOnCancel(ctx, cancelled.BapUri, onCancelReq)
}

func (s *Service) restoreStock(ctx context.Context, q *dbsqlc.Queries, contractID uuid.UUID) {
	resourceIDs, err := q.GetCommitmentResourceIDsByContractID(ctx, contractID)
	if err != nil {
		s.lh.WithModule("cancelsvc").Warn().Error(err).Log("could not load resource IDs for stock restore")
		return
	}
	seen := map[string]struct{}{}
	for _, rid := range resourceIDs {
		if _, ok := seen[rid]; ok {
			continue
		}
		seen[rid] = struct{}{}
		// Restore: set quantity back up by 1. We use UpsertResourceStock with a
		// relative increment via a raw query approach. For simplicity, we add 1 back
		// by doing an UPDATE directly; the generated code doesn't have an increment
		// query, so we call the pool directly.
		_, err := s.pool.Exec(ctx,
			`UPDATE resource_stock SET quantity = quantity + 1, sold = GREATEST(0, sold - 1), updated_at = NOW()
			 WHERE resource_id = $1 AND bpp_id = $2`,
			rid, s.cfg.BppID)
		if err != nil {
			s.lh.WithModule("cancelsvc").Warn().Error(err).Log("stock restore failed for resource " + rid)
		} else {
			log.Printf("[stock] restored stock for resource=%s bpp=%s", rid, s.cfg.BppID)
		}
	}
}

func (s *Service) callOnCancel(ctx context.Context, bapURI string, req OnCancelRequest) {
	url := bapURI + "/on_cancel"
	body, _ := json.Marshal(req)
	prettyLog(fmt.Sprintf("→ OUT  ON_CANCEL | txn=%.8s… | POST %s", req.Context.TransactionID, url), req)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		s.lh.WithModule("cancelsvc").Err().Error(err).Log("build on_cancel request failed")
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		s.lh.WithModule("cancelsvc").Warn().Error(err).Log("on_cancel callback failed")
		return
	}
	defer resp.Body.Close() //nolint:errcheck
	respBody, _ := io.ReadAll(resp.Body)
	prettyLog(fmt.Sprintf("← ACK  ON_CANCEL | http=%d", resp.StatusCode), map[string]interface{}{"response": json.RawMessage(respBody)})
}

func prettyLog(label string, v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Printf("[beckn] %s\n(marshal error: %v)", label, err)
		return
	}
	log.Printf("[beckn] %s\n%s", label, string(b))
}
