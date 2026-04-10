package initsvc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/remiges-tech/logharbour/logharbour"

	dbsqlc "github.com/ion/winroom/bpp/db/sqlc"
	"github.com/ion/winroom/bpp/internal/config"
)

// InitService handles the full /beckn/init flow:
//  1. Persist the contract state (participants, commitments, considerations, performances)
//     to the database — this is the "add-to-cart" equivalent in the Beckn lifecycle.
//  2. Build the on_init payload echoing back the enriched contract with its assigned id.
//  3. POST the on_init callback to the BAP.
//  4. Write INBOUND and OUTBOUND audit rows to beckn_message_log.
type InitService struct {
	pool       *pgxpool.Pool
	cfg        *config.Config
	lh         *logharbour.Logger
	httpClient *http.Client
}

func NewInitService(pool *pgxpool.Pool, cfg *config.Config, lh *logharbour.Logger) *InitService {
	return &InitService{
		pool:       pool,
		cfg:        cfg,
		lh:         lh,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// ProcessInit is called asynchronously after the BPP returns ACK to the BAP.
// It persists the contract state, calls on_init on the BAP, and logs both messages.
func (s *InitService) ProcessInit(ctx context.Context, req *InitRequest) {
	start := time.Now()

	txID, _ := uuid.Parse(req.Context.TransactionID)
	inboundMsgID, _ := uuid.Parse(req.Context.MessageID)
	reqJSON, _ := json.Marshal(req)

	// Persist contract state and get (or create) the contract ID.
	contractID, err := s.persistContract(ctx, req)
	if err != nil {
		s.lh.WithModule("initsvc").Err().Error(err).Log("failed to persist contract state")
		// Still attempt to log INBOUND even on DB failure.
	}

	// Log INBOUND.
	s.writeMessageLog(ctx, dbsqlc.InsertMessageLogParams{
		TransactionID:        uuidToPgtype(txID),
		MessageID:            inboundMsgID,
		Action:               dbsqlc.BecknActionInit,
		Direction:            dbsqlc.MessageDirectionINBOUND,
		Url:                  strPtr("/webhook/init"),
		BapID:                strPtr(req.Context.BapID),
		BapUri:               strPtr(req.Context.BapURI),
		BppID:                strPtr(req.Context.BppID),
		BppUri:               strPtr(req.Context.BppURI),
		NetworkID:            strPtr(req.Context.NetworkID),
		AckStatus:            dbsqlc.NullAckStatus{AckStatus: dbsqlc.AckStatusACK, Valid: true},
		RequestPayload:       reqJSON,
		ProcessingDurationMs: durationMs(start),
	})

	if err != nil {
		return
	}

	// Build the on_init payload — echo the contract back with the assigned id.
	outboundMsgID := uuid.New()
	responseContract := req.Message.Contract
	responseContract.ID = contractID.String()

	onInitReq := OnInitRequest{
		Context: BecknContext{
			Version:       req.Context.Version,
			Action:        "on_init",
			Timestamp:     time.Now().UTC().Format(time.RFC3339),
			MessageID:     outboundMsgID.String(),
			TransactionID: req.Context.TransactionID,
			BapID:         req.Context.BapID,
			BapURI:        req.Context.BapURI,
			BppID:         req.Context.BppID,
			BppURI:        req.Context.BppURI,
			TTL:           req.Context.TTL,
			NetworkID:     req.Context.NetworkID,
		},
		Message: OnInitMessage{Contract: responseContract},
	}

	// POST on_init to BAP and log OUTBOUND.
	s.callOnInit(ctx, txID, outboundMsgID, req.Context.BapURI, onInitReq)
}

// persistContract writes all contract sub-entities to the DB in a single transaction.
// Returns the UUID of the contract row (new or existing).
func (s *InitService) persistContract(ctx context.Context, req *InitRequest) (uuid.UUID, error) {
	q := dbsqlc.New(s.pool)
	bppID := s.cfg.BppID

	txID, err := uuid.Parse(req.Context.TransactionID)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid transactionId: %w", err)
	}

	// Look up existing DRAFT contract for this transaction, or create one.
	contractID, lookupErr := q.GetContractByTransactionID(ctx, dbsqlc.GetContractByTransactionIDParams{
		TransactionID: txID,
		BppID:         bppID,
	})

	if lookupErr != nil && !errors.Is(lookupErr, pgx.ErrNoRows) {
		return uuid.UUID{}, fmt.Errorf("contract lookup failed: %w", lookupErr)
	}

	// Extract contractAttributes JSON-LD context/type if present.
	caCtx, caType := extractJSONLDMeta(req.Message.Contract.ContractAttributes)
	caJSON := req.Message.Contract.ContractAttributes
	if len(caJSON) == 0 {
		caJSON = json.RawMessage("null")
	}

	if errors.Is(lookupErr, pgx.ErrNoRows) {
		// No existing contract — create a fresh one.
		// If the BAP supplied an ID use it; otherwise generate one.
		newID := uuid.New()
		if req.Message.Contract.ID != "" {
			if parsed, err := uuid.Parse(req.Message.Contract.ID); err == nil {
				newID = parsed
			}
		}
		contractID, err = q.InsertContract(ctx, dbsqlc.InsertContractParams{
			ID:                        newID,
			BppID:                     bppID,
			BapID:                     req.Context.BapID,
			BapUri:                    req.Context.BapURI,
			TransactionID:             txID,
			NetworkID:                 strPtr(req.Context.NetworkID),
			ContractAttributes:        caJSON,
			ContractAttributesContext: caCtx,
			ContractAttributesType:    caType,
		})
		if err != nil {
			return uuid.UUID{}, fmt.Errorf("insert contract: %w", err)
		}
	} else {
		// Update the existing contract's mutable fields.
		if err := q.UpdateContractAttributes(ctx, dbsqlc.UpdateContractAttributesParams{
			ID:                        contractID,
			BapUri:                    req.Context.BapURI,
			ContractAttributes:        caJSON,
			ContractAttributesContext: caCtx,
			ContractAttributesType:    caType,
		}); err != nil {
			return uuid.UUID{}, fmt.Errorf("update contract attributes: %w", err)
		}
	}

	// Persist participants.
	for _, p := range req.Message.Contract.Participants {
		paCtx, paType := extractJSONLDMeta(p.ParticipantAttributes)
		paJSON := p.ParticipantAttributes
		if len(paJSON) == 0 {
			paJSON = json.RawMessage("null")
		}
		descName := ""
		if p.Descriptor != nil {
			descName = p.Descriptor.Name
		}
		if err := q.UpsertParticipant(ctx, dbsqlc.UpsertParticipantParams{
			ID:                           p.ID,
			ContractID:                   contractID,
			DescriptorName:               strPtr(descName),
			ParticipantAttributes:        paJSON,
			ParticipantAttributesContext: paCtx,
			ParticipantAttributesType:    paType,
		}); err != nil {
			s.lh.WithModule("initsvc").Warn().Error(err).Log(fmt.Sprintf("upsert participant %q failed", p.ID))
		}
	}

	// Persist commitments, their offers, and link resources.
	for _, c := range req.Message.Contract.Commitments {
		commAttrJSON := c.CommitmentAttributes
		if len(commAttrJSON) == 0 {
			commAttrJSON = json.RawMessage("null")
		}
		commCtx, commType := extractJSONLDMeta(commAttrJSON)
		if err := q.UpsertCommitment(ctx, dbsqlc.UpsertCommitmentParams{
			ID:                          c.ID,
			ContractID:                  contractID,
			CommitmentAttributes:        commAttrJSON,
			CommitmentAttributesContext: commCtx,
			CommitmentAttributesType:    commType,
		}); err != nil {
			s.lh.WithModule("initsvc").Warn().Error(err).Log(fmt.Sprintf("upsert commitment %q failed", c.ID))
			continue
		}

		// Persist the offer if present.
		if c.Offer != nil {
			offerDescJSON, _ := json.Marshal(c.Offer.Descriptor)
			offerAttrJSON := c.Offer.OfferAttributes
			if len(offerAttrJSON) == 0 {
				offerAttrJSON = json.RawMessage("null")
			}
			resourceIDs := c.Offer.ResourceIDs
			if resourceIDs == nil {
				resourceIDs = []string{}
			}
			if err := q.UpsertCommitmentOffer(ctx, dbsqlc.UpsertCommitmentOfferParams{
				CommitmentID:    c.ID,
				ContractID:      contractID,
				OfferID:         c.Offer.ID,
				OfferDescriptor: offerDescJSON,
				OfferAttributes: offerAttrJSON,
				ResourceIds:     resourceIDs,
			}); err != nil {
				s.lh.WithModule("initsvc").Warn().Error(err).Log(fmt.Sprintf("upsert commitment offer %q failed", c.Offer.ID))
			}
		}
	}

	// Persist considerations.
	for _, con := range req.Message.Contract.Consideration {
		statusCode := "QUOTED"
		if con.Status != nil {
			statusCode = con.Status.Code
		}
		conAttrJSON := con.ConsiderationAttributes
		if len(conAttrJSON) == 0 {
			conAttrJSON = json.RawMessage("null")
		}
		conCtx, conType := extractJSONLDMeta(conAttrJSON)
		if err := q.UpsertConsideration(ctx, dbsqlc.UpsertConsiderationParams{
			ID:                             con.ID,
			ContractID:                     contractID,
			StatusCode:                     statusCode,
			ConsiderationAttributes:        conAttrJSON,
			ConsiderationAttributesContext: conCtx,
			ConsiderationAttributesType:    conType,
		}); err != nil {
			s.lh.WithModule("initsvc").Warn().Error(err).Log(fmt.Sprintf("upsert consideration %q failed", con.ID))
		}
	}

	// Persist performances (fulfillment units).
	for _, perf := range req.Message.Contract.Performance {
		statusCode := ""
		if perf.Status != nil {
			statusCode = perf.Status.Code
		}
		perfAttrJSON := perf.PerformanceAttributes
		if len(perfAttrJSON) == 0 {
			perfAttrJSON = json.RawMessage("null")
		}
		perfCtx, perfType := extractJSONLDMeta(perfAttrJSON)
		commitmentIDs := perf.CommitmentIDs
		if commitmentIDs == nil {
			commitmentIDs = []string{}
		}
		if err := q.UpsertPerformance(ctx, dbsqlc.UpsertPerformanceParams{
			ID:                           perf.ID,
			ContractID:                   contractID,
			StatusCode:                   strPtr(statusCode),
			CommitmentIds:                commitmentIDs,
			PerformanceAttributes:        perfAttrJSON,
			PerformanceAttributesContext: perfCtx,
			PerformanceAttributesType:    perfType,
		}); err != nil {
			s.lh.WithModule("initsvc").Warn().Error(err).Log(fmt.Sprintf("upsert performance %q failed", perf.ID))
		}
	}

	return contractID, nil
}

// callOnInit POSTs the on_init payload to the BAP and logs the OUTBOUND entry.
func (s *InitService) callOnInit(
	ctx context.Context,
	txID uuid.UUID,
	msgID uuid.UUID,
	bapURI string,
	req OnInitRequest,
) {
	start := time.Now()
	onInitURL := bapURI + "/on_init"
	reqJSON, _ := json.Marshal(req)

	var (
		respBody []byte
		callErr  error
	)

	defer func() {
		ackStatus := dbsqlc.NullAckStatus{AckStatus: dbsqlc.AckStatusACK, Valid: true}
		if callErr != nil {
			ackStatus = dbsqlc.NullAckStatus{AckStatus: dbsqlc.AckStatusNACK, Valid: true}
		}
		s.writeMessageLog(ctx, dbsqlc.InsertMessageLogParams{
			TransactionID:        uuidToPgtype(txID),
			MessageID:            msgID,
			Action:               dbsqlc.BecknActionOnInit,
			Direction:            dbsqlc.MessageDirectionOUTBOUND,
			Url:                  strPtr(onInitURL),
			BapID:                strPtr(req.Context.BapID),
			BapUri:               strPtr(req.Context.BapURI),
			BppID:                strPtr(req.Context.BppID),
			BppUri:               strPtr(req.Context.BppURI),
			NetworkID:            strPtr(req.Context.NetworkID),
			AckStatus:            ackStatus,
			RequestPayload:       reqJSON,
			ResponsePayload:      rawMessageOrNull(respBody),
			ErrorMessage:         errStrPtr(callErr),
			ProcessingDurationMs: durationMs(start),
		})
	}()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, onInitURL, bytes.NewReader(reqJSON))
	if err != nil {
		callErr = fmt.Errorf("build on_init request: %w", err)
		s.lh.WithModule("initsvc").Err().Error(callErr).Log("on_init HTTP request build failed")
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		callErr = fmt.Errorf("on_init HTTP call to BAP failed: %w", err)
		s.lh.WithModule("initsvc").Err().Error(callErr).Log("on_init call failed")
		return
	}
	defer resp.Body.Close() //nolint:errcheck

	respBody, _ = io.ReadAll(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		callErr = fmt.Errorf("BAP returned non-2xx status %d: %s", resp.StatusCode, string(respBody))
		s.lh.WithModule("initsvc").Warn().Error(callErr).Log("on_init BAP response error")
		return
	}

	s.lh.WithModule("initsvc").Log("on_init sent to BAP successfully")
}

// writeMessageLog inserts one audit row into beckn_message_log.
func (s *InitService) writeMessageLog(ctx context.Context, params dbsqlc.InsertMessageLogParams) {
	q := dbsqlc.New(s.pool)
	if err := q.InsertMessageLog(ctx, params); err != nil {
		s.lh.WithModule("initsvc").Warn().Error(err).Log("failed to write beckn_message_log")
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// extractJSONLDMeta pulls @context and @type from a JSON-LD blob.
// Returns nil pointers if the fields are absent so sqlc can store NULL.
func extractJSONLDMeta(raw json.RawMessage) (ctx *string, typ *string) {
	if len(raw) == 0 {
		return nil, nil
	}
	var m struct {
		Context string `json:"@context"`
		Type    string `json:"@type"`
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, nil
	}
	if m.Context != "" {
		ctx = &m.Context
	}
	if m.Type != "" {
		typ = &m.Type
	}
	return ctx, typ
}

// rawMessageOrNull returns nil (SQL NULL) when b is empty so that
// pgx does not send an empty byte slice to a JSONB column, which
// Postgres rejects as invalid JSON.
func rawMessageOrNull(b []byte) json.RawMessage {
	if len(b) == 0 {
		return nil
	}
	return json.RawMessage(b)
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func errStrPtr(err error) *string {
	if err == nil {
		return nil
	}
	s := err.Error()
	return &s
}

func uuidToPgtype(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func durationMs(start time.Time) *int32 {
	ms := int32(time.Since(start).Milliseconds())
	return &ms
}
