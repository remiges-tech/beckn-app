package statussvc

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/remiges-tech/logharbour/logharbour"

	"github.com/ion/winroom/bpp/internal/config"
)

type Handler struct {
	svc *Service
	lh  *logharbour.Logger
}

func NewHandler(pool *pgxpool.Pool, cfg *config.Config, lh *logharbour.Logger) *Handler {
	return &Handler{svc: NewService(pool, cfg, lh), lh: lh}
}

func (h *Handler) Handle(c *gin.Context) {
	var req StatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.lh.WithModule("statussvc").Err().Error(err).Log("invalid status request")
		c.JSON(http.StatusBadRequest, buildNACK(BecknContext{}, "invalid request body"))
		return
	}
	if req.Context.BapURI == "" {
		c.JSON(http.StatusBadRequest, buildNACK(req.Context, "bapUri required"))
		return
	}
	c.JSON(http.StatusOK, BecknACK{Context: ackCtx(req.Context), Message: ACKMessage{ACK: ACKStatus{Status: "ACK"}}})
	go func() { h.svc.Process(context.Background(), &req) }()
}

func ackCtx(in BecknContext) BecknContext { return in }

func buildNACK(ctx BecknContext, _ string) BecknACK {
	return BecknACK{Context: ackCtx(ctx), Message: ACKMessage{ACK: ACKStatus{Status: "NACK"}}}
}
