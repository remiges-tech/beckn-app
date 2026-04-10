package statussvc

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

// Process looks up the current contract state and fires on_status to the BAP.
func (s *Service) Process(ctx context.Context, req *StatusRequest) {
	prettyLog(fmt.Sprintf("← IN   STATUS   | txn=%.8s… | bap=%s", req.Context.TransactionID, req.Context.BapID), req)

	txID, err := uuid.Parse(req.Context.TransactionID)
	if err != nil {
		s.lh.WithModule("statussvc").Err().Error(err).Log("invalid transactionId")
		return
	}

	q := dbsqlc.New(s.pool)
	row, err := q.GetActiveContractByTxnID(ctx, dbsqlc.GetActiveContractByTxnIDParams{
		TransactionID: txID,
		BppID:         s.cfg.BppID,
	})
	if err != nil {
		s.lh.WithModule("statussvc").Err().Error(err).Log("contract not found for status")
		return
	}

	// Build a minimal contract payload reflecting current status.
	contract := map[string]interface{}{
		"id":     row.ID.String(),
		"status": map[string]string{"code": string(row.Status)},
	}
	if row.ContractAttributes != nil {
		contract["contractAttributes"] = json.RawMessage(row.ContractAttributes)
	}
	contractJSON, _ := json.Marshal(contract)

	outMsgID := uuid.New()
	onStatusReq := OnStatusRequest{
		Context: BecknContext{
			Version:       req.Context.Version,
			Action:        "on_status",
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
		Message: OnStatusMessage{Contract: contractJSON},
	}

	s.callOnStatus(ctx, req.Context.BapURI, onStatusReq)
}

func (s *Service) callOnStatus(ctx context.Context, bapURI string, req OnStatusRequest) {
	start := time.Now()
	url := bapURI + "/on_status"
	body, _ := json.Marshal(req)

	prettyLog(fmt.Sprintf("→ OUT  ON_STATUS | txn=%.8s… | POST %s", req.Context.TransactionID, url), req)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		s.lh.WithModule("statussvc").Err().Error(err).Log("build on_status request failed")
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		prettyLog(fmt.Sprintf("← ERR  ON_STATUS | dur=%dms | %s", time.Since(start).Milliseconds(), err), map[string]string{"error": err.Error()})
		return
	}
	defer resp.Body.Close() //nolint:errcheck
	respBody, _ := io.ReadAll(resp.Body)
	prettyLog(fmt.Sprintf("← ACK  ON_STATUS | http=%d | dur=%dms", resp.StatusCode, time.Since(start).Milliseconds()), map[string]interface{}{"response": json.RawMessage(respBody)})
}

func prettyLog(label string, v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Printf("[beckn] %s\n(marshal error: %v)", label, err)
		return
	}
	log.Printf("[beckn] %s\n%s", label, string(b))
}
