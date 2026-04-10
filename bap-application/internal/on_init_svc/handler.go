package oninitsvc

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/remiges-tech/logharbour/logharbour"

	"github.com/ion/winroom/bap/internal/config"
)

// OnInitHandler handles POST /bap/receiver/on_init — a callback from the BPP
// carrying the draft contract with a BPP-assigned contract id.
type OnInitHandler struct {
	svc *OnInitService
	lh  *logharbour.Logger
}

func NewOnInitHandler(pool *pgxpool.Pool, cfg *config.Config, lh *logharbour.Logger) *OnInitHandler {
	return &OnInitHandler{
		svc: NewOnInitService(pool, cfg, lh),
		lh:  lh,
	}
}

func (h *OnInitHandler) Handle(c *gin.Context) {
	var req OnInitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.lh.WithModule("on_init_handler").Err().Error(err).Log("invalid on_init request body")
		c.JSON(http.StatusBadRequest, buildNACK(BecknContext{}, err.Error()))
		return
	}

	c.JSON(http.StatusOK, BecknACK{
		Context: contextForACK(req.Context, "on_init"),
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
		Context: contextForACK(ctx, "on_init"),
		Message: ACKMessage{ACK: ACKStatus{Status: "NACK"}},
	}
}
