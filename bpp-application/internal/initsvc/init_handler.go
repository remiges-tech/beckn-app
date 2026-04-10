package initsvc

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/remiges-tech/logharbour/logharbour"

	"github.com/ion/winroom/bpp/internal/config"
)

// InitHandler handles POST /beckn/init inbound from the BAP.
// Beckn protocol requires an immediate ACK response; the actual on_init
// callback is made asynchronously so the BAP's connection is not blocked.
// Unlike select, init also persists the contract state to the database.
type InitHandler struct {
	svc *InitService
	lh  *logharbour.Logger
}

func NewInitHandler(pool *pgxpool.Pool, cfg *config.Config, lh *logharbour.Logger) *InitHandler {
	return &InitHandler{
		svc: NewInitService(pool, cfg, lh),
		lh:  lh,
	}
}

// Handle is the Gin handler for POST /beckn/init.
func (h *InitHandler) Handle(c *gin.Context) {
	var req InitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.lh.WithModule("init_handler").Err().Error(err).Log("invalid init request body")
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

	// Process the init and call on_init asynchronously.
	// context.Background() is used deliberately — c.Request.Context() is cancelled
	// by Gin as soon as the ACK response above is flushed, which would abort
	// every downstream DB write and the on_init HTTP call.
	go func() {
		h.svc.ProcessInit(context.Background(), &req)
	}()
}

// contextForACK returns a minimal context suitable for the synchronous ACK.
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
