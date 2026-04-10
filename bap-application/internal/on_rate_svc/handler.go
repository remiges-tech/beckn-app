package onratesvc

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/remiges-tech/logharbour/logharbour"

	"github.com/ion/winroom/bap/internal/config"
)

type Handler struct {
	svc *Service
	lh  *logharbour.Logger
}

func NewHandler(pool *pgxpool.Pool, cfg *config.Config, lh *logharbour.Logger) *Handler {
	return &Handler{svc: NewService(pool, cfg, lh), lh: lh}
}

func (h *Handler) Handle(c *gin.Context) {
	var req OnRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.lh.WithModule("on_rate_handler").Err().Error(err).Log("invalid on_rate body")
		c.JSON(http.StatusBadRequest, buildNACK(BecknContext{}, err.Error()))
		return
	}
	c.JSON(http.StatusOK, BecknACK{
		Context: ackCtx(req.Context),
		Message: ACKMessage{ACK: ACKStatus{Status: "ACK"}},
	})
	go func() { h.svc.Process(context.Background(), &req) }()
}

func ackCtx(in BecknContext) BecknContext {
	out := in
	out.Action = "on_rate"
	return out
}

func buildNACK(ctx BecknContext, _ string) BecknACK {
	return BecknACK{Context: ackCtx(ctx), Message: ACKMessage{ACK: ACKStatus{Status: "NACK"}}}
}
