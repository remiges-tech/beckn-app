package confirmsvc

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/remiges-tech/logharbour/logharbour"

	"github.com/ion/winroom/bpp/internal/config"
)

// ConfirmHandler handles POST /beckn/confirm inbound from the BAP.
// Beckn protocol requires an immediate ACK response; the actual on_confirm
// callback is made asynchronously so the BAP's connection is not blocked.
// Confirm transitions the contract from DRAFT → ACTIVE in the database.
type ConfirmHandler struct {
	svc *ConfirmService
	lh  *logharbour.Logger
}

func NewConfirmHandler(pool *pgxpool.Pool, cfg *config.Config, lh *logharbour.Logger) *ConfirmHandler {
	return &ConfirmHandler{
		svc: NewConfirmService(pool, cfg, lh),
		lh:  lh,
	}
}

// Handle is the Gin handler for POST /beckn/confirm.
func (h *ConfirmHandler) Handle(c *gin.Context) {
	var req ConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.lh.WithModule("confirm_handler").Err().Error(err).Log("invalid confirm request body")
		c.JSON(http.StatusBadRequest, buildNACK(BecknContext{}, "invalid request body"))
		return
	}

	if req.Context.BapURI == "" {
		c.JSON(http.StatusBadRequest, buildNACK(req.Context, "bapUri is required in context"))
		return
	}

	// Return ACK immediately — Beckn spec requires a synchronous ACK within TTL.
	c.JSON(http.StatusOK, BecknACK{
		Context: contextForACK(req.Context),
		Message: ACKMessage{ACK: ACKStatus{Status: "ACK"}},
	})

	// Process the confirm and call on_confirm asynchronously.
	// context.Background() is used deliberately — c.Request.Context() is cancelled
	// by Gin as soon as the ACK response above is flushed.
	go func() {
		h.svc.ProcessConfirm(context.Background(), &req)
	}()
}

func contextForACK(in BecknContext) BecknContext {
	return BecknContext{
		Version:       in.Version,
		Action:        in.Action,
		Timestamp:     in.Timestamp,
		MessageID:     in.MessageID,
		TransactionID: in.TransactionID,
		BapID:         in.BapID,
		BapURI:        in.BapURI,
		BppID:         in.BppID,
		BppURI:        in.BppURI,
		TTL:           in.TTL,
		NetworkID:     in.NetworkID,
	}
}

func buildNACK(ctx BecknContext, reason string) BecknACK {
	_ = reason
	return BecknACK{
		Context: contextForACK(ctx),
		Message: ACKMessage{ACK: ACKStatus{Status: "NACK"}},
	}
}
