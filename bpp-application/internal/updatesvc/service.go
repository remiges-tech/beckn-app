package updatesvc

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

// Process applies the partial contract update and fires on_update to the BAP.
func (s *Service) Process(ctx context.Context, req *UpdateRequest) {
	prettyLog(fmt.Sprintf("← IN   UPDATE   | txn=%.8s… | bap=%s", req.Context.TransactionID, req.Context.BapID), req)

	txID, err := uuid.Parse(req.Context.TransactionID)
	if err != nil {
		s.lh.WithModule("updatesvc").Err().Error(err).Log("invalid transactionId")
		return
	}

	q := dbsqlc.New(s.pool)
	row, err := q.GetActiveContractByTxnID(ctx, dbsqlc.GetActiveContractByTxnIDParams{
		TransactionID: txID,
		BppID:         s.cfg.BppID,
	})
	if err != nil {
		s.lh.WithModule("updatesvc").Err().Error(err).Log("contract not found for update")
		return
	}

	// Persist the update payload as new contractAttributes.
	if req.Message.Contract != nil {
		_, err = s.pool.Exec(ctx,
			`UPDATE contracts SET contract_attributes = $1, updated_at = NOW()
			 WHERE id = $2 AND bpp_id = $3`,
			[]byte(req.Message.Contract), row.ID, s.cfg.BppID)
		if err != nil {
			s.lh.WithModule("updatesvc").Err().Error(err).Log("update contract attributes failed")
			return
		}
	}

	// Echo back the updated contract in on_update.
	updatedContract := req.Message.Contract
	if updatedContract == nil {
		updatedContract = json.RawMessage(`{"id":"` + row.ID.String() + `"}`)
	}

	outMsgID := uuid.New()
	onUpdateReq := OnUpdateRequest{
		Context: BecknContext{
			Version:       req.Context.Version,
			Action:        "on_update",
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
		Message: OnUpdateMessage{Contract: updatedContract},
	}

	s.callOnUpdate(ctx, row.BapUri, onUpdateReq)
}

func (s *Service) callOnUpdate(ctx context.Context, bapURI string, req OnUpdateRequest) {
	url := bapURI + "/on_update"
	body, _ := json.Marshal(req)
	prettyLog(fmt.Sprintf("→ OUT  ON_UPDATE | txn=%.8s… | POST %s", req.Context.TransactionID, url), req)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		s.lh.WithModule("updatesvc").Err().Error(err).Log("build on_update request failed")
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		s.lh.WithModule("updatesvc").Warn().Error(err).Log("on_update callback failed")
		return
	}
	defer resp.Body.Close() //nolint:errcheck
	respBody, _ := io.ReadAll(resp.Body)
	prettyLog(fmt.Sprintf("← ACK  ON_UPDATE | http=%d", resp.StatusCode), map[string]interface{}{"response": json.RawMessage(respBody)})
}

func prettyLog(label string, v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Printf("[beckn] %s\n(marshal error: %v)", label, err)
		return
	}
	log.Printf("[beckn] %s\n%s", label, string(b))
}
