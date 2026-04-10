package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/remiges-tech/logharbour/logharbour"

	"github.com/ion/winroom/bap/internal/clientsvc"
	"github.com/ion/winroom/bap/internal/config"
	oncancelsvc "github.com/ion/winroom/bap/internal/on_cancel_svc"
	onconfirmsvc "github.com/ion/winroom/bap/internal/on_confirm_svc"
	oninitsvc "github.com/ion/winroom/bap/internal/on_init_svc"
	onratesvc "github.com/ion/winroom/bap/internal/on_rate_svc"
	onselectsvc "github.com/ion/winroom/bap/internal/on_select_svc"
	onstatussvc "github.com/ion/winroom/bap/internal/on_status_svc"
	onsupportsvc "github.com/ion/winroom/bap/internal/on_support_svc"
	onupdatesvc "github.com/ion/winroom/bap/internal/on_update_svc"
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

		onStatusH := onstatussvc.NewHandler(pool, cfg, lh)
		webhook.POST("/on_status", onStatusH.Handle)

		onCancelH := oncancelsvc.NewHandler(pool, cfg, lh)
		webhook.POST("/on_cancel", onCancelH.Handle)

		onUpdateH := onupdatesvc.NewHandler(pool, cfg, lh)
		webhook.POST("/on_update", onUpdateH.Handle)

		onRateH := onratesvc.NewHandler(pool, cfg, lh)
		webhook.POST("/on_rate", onRateH.Handle)

		onSupportH := onsupportsvc.NewHandler(pool, cfg, lh)
		webhook.POST("/on_support", onSupportH.Handle)
	}

	// Client-facing APIs for the React frontend
	v1 := r.Group("/api/v1")
	{
		clientH := clientsvc.NewClientHandler(pool, cfg, lh)
		v1.GET("/catalog", clientH.HandleCatalog)
		v1.GET("/discover", clientH.HandleDiscover)
		v1.POST("/select", clientH.HandleSelect)
		v1.POST("/init", clientH.HandleInit)
		v1.POST("/confirm", clientH.HandleConfirm)
		v1.GET("/status/:id", clientH.HandleStatus)
		v1.POST("/request-status", clientH.HandleRequestStatus)
		v1.POST("/cancel", clientH.HandleCancel)
		v1.POST("/update", clientH.HandleUpdate)
		v1.POST("/rate", clientH.HandleRate)
		v1.POST("/support", clientH.HandleSupport)
	}
}
