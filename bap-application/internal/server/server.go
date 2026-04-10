package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/remiges-tech/logharbour/logharbour"

	"github.com/ion/winroom/bap/internal/config"
	onconfirmsvc "github.com/ion/winroom/bap/internal/on_confirm_svc"
	oninitsvc "github.com/ion/winroom/bap/internal/on_init_svc"
	onselectsvc "github.com/ion/winroom/bap/internal/on_select_svc"
)

// RegisterRoutes wires all BAP routes onto the given Gin engine.
func RegisterRoutes(r *gin.Engine, pool *pgxpool.Pool, cfg *config.Config, lh *logharbour.Logger) {
	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Beckn webhook endpoints — BPP (via ONIX adapter) POSTs on_* callbacks here.
	webhook := r.Group("/api/webhook")
	{
		onSelectH := onselectsvc.NewOnSelectHandler(pool, cfg, lh)
		webhook.POST("/on_select", onSelectH.Handle)

		onInitH := oninitsvc.NewOnInitHandler(pool, cfg, lh)
		webhook.POST("/on_init", onInitH.Handle)

		onConfirmH := onconfirmsvc.NewOnConfirmHandler(pool, cfg, lh)
		webhook.POST("/on_confirm", onConfirmH.Handle)
	}
}
