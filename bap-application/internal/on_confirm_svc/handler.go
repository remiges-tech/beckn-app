package onconfirmsvc

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/remiges-tech/logharbour/logharbour"

	"github.com/ion/winroom/bap/internal/config"
)

// OnConfirmHandler handles POST /bap/receiver/on_confirm — a callback from the BPP
// with the confirmed (ACTIVE) contract.
type OnConfirmHandler struct {
	svc *OnConfirmService
	lh  *logharbour.Logger
}

func NewOnConfirmHandler(pool *pgxpool.Pool, cfg *config.Config, lh *logharbour.Logger) *OnConfirmHandler {
	return &OnConfirmHandler{
		svc: NewOnConfirmService(pool, cfg, lh),
		lh:  lh,
	}
}

func (h *OnConfirmHandler) Handle(c *gin.Context) {
	var req OnConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.lh.WithModule("on_confirm_handler").Err().Error(err).Log("invalid on_confirm request body")
		c.JSON(http.StatusBadRequest, buildNACK(BecknContext{}, err.Error()))
		return
	}

	c.JSON(http.StatusOK, BecknACK{
		Context: contextForACK(req.Context, "on_confirm"),
		Message: ACKMessage{ACK: ACKStatus{Status: "ACK"}},
	})

	go func() {
		h.svc.Process(context.Background(), &req)
	}()
}

func contextForACK(in BecknContext, action string) BecknContext {
	out := in
	out.Action = action
	return out
}

func buildNACK(ctx BecknContext, _ string) BecknACK {
	return BecknACK{
		Context: contextForACK(ctx, "on_confirm"),
		Message: ACKMessage{ACK: ACKStatus{Status: "NACK"}},
	}
}
