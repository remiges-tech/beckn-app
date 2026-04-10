package supportsvc

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
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/remiges-tech/logharbour/logharbour"

	dbsqlc "github.com/ion/winroom/bpp/db/sqlc"
	"github.com/ion/winroom/bpp/internal/config"
)

// defaultChannels is the BPP's configured support contact list returned to BAPs.
// In production this should come from config.
var defaultChannels = json.RawMessage(`[{"type":"email","contact":"support@bpp.example.com"}]`)

type Service struct {
	pool       *pgxpool.Pool
	cfg        *config.Config
	lh         *logharbour.Logger
	httpClient *http.Client
}

func NewService(pool *pgxpool.Pool, cfg *config.Config, lh *logharbour.Logger) *Service {
	return &Service{pool: pool, cfg: cfg, lh: lh, httpClient: &http.Client{Timeout: 30 * time.Second}}
}

// Process creates a support ticket in the DB and fires on_support to the BAP.
func (s *Service) Process(ctx context.Context, req *SupportRequest) {
	prettyLog(fmt.Sprintf("← IN   SUPPORT  | txn=%.8s… | bap=%s", req.Context.TransactionID, req.Context.BapID), req)

	q := dbsqlc.New(s.pool)

	// Resolve the contract ID from the order ID if provided.
	var contractID pgtype.UUID
	if req.Message.Support.OrderID != "" {
		txID, err := uuid.Parse(req.Message.Support.OrderID)
		if err == nil {
			row, err := q.GetActiveContractByTxnID(ctx, dbsqlc.GetActiveContractByTxnIDParams{
				TransactionID: txID,
				BppID:         s.cfg.BppID,
			})
			if err == nil {
				contractID = pgtype.UUID{Bytes: row.ID, Valid: true}
			}
		}
	}

	channels := req.Message.Support.Channels
	if channels == nil {
		channels = defaultChannels
	}

	descName := &req.Message.Support.Descriptor.Name
	if *descName == "" {
		descName = nil
	}
	descShort := &req.Message.Support.Descriptor.ShortDesc
	if *descShort == "" {
		descShort = nil
	}

	ticketID, err := q.InsertSupportTicket(ctx, dbsqlc.InsertSupportTicketParams{
		ContractID:          contractID,
		BapID:               req.Context.BapID,
		BapUri:              req.Context.BapURI,
		DescriptorName:      descName,
		DescriptorShortDesc: descShort,
		Channels:            defaultChannels,
		IsPreview:           false,
	})
	if err != nil {
		s.lh.WithModule("supportsvc").Err().Error(err).Log("insert support ticket failed")
		return
	}
	log.Printf("[support] created ticket=%s txn=%.8s…", ticketID, req.Context.TransactionID)

	outMsgID := uuid.New()
	onSupportReq := OnSupportRequest{
		Context: BecknContext{
			Version:       req.Context.Version,
			Action:        "on_support",
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
		Message: OnSupportMessage{
			Support: OnSupportPayload{
				TicketID: ticketID.String(),
				Channels: defaultChannels,
			},
		},
	}
	s.callOnSupport(ctx, req.Context.BapURI, onSupportReq)
}

func (s *Service) callOnSupport(ctx context.Context, bapURI string, req OnSupportRequest) {
	url := bapURI + "/on_support"
	body, _ := json.Marshal(req)
	prettyLog(fmt.Sprintf("→ OUT  ON_SUPPORT| txn=%.8s… | POST %s", req.Context.TransactionID, url), req)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		s.lh.WithModule("supportsvc").Err().Error(err).Log("build on_support request failed")
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		s.lh.WithModule("supportsvc").Warn().Error(err).Log("on_support callback failed")
		return
	}
	defer resp.Body.Close() //nolint:errcheck
	respBody, _ := io.ReadAll(resp.Body)
	prettyLog(fmt.Sprintf("← ACK  ON_SUPPORT| http=%d", resp.StatusCode), map[string]interface{}{"response": json.RawMessage(respBody)})
}

func prettyLog(label string, v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Printf("[beckn] %s\n(marshal error: %v)", label, err)
		return
	}
	log.Printf("[beckn] %s\n%s", label, string(b))
}
