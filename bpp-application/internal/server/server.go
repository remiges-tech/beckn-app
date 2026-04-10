package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/remiges-tech/alya/service"
	"github.com/remiges-tech/alya/wscutils"

	"github.com/ion/winroom/bpp/internal/cancelsvc"
	"github.com/ion/winroom/bpp/internal/catalog"
	"github.com/ion/winroom/bpp/internal/config"
	"github.com/ion/winroom/bpp/internal/confirmsvc"
	"github.com/ion/winroom/bpp/internal/dashboardsvc"
	"github.com/ion/winroom/bpp/internal/initsvc"
	"github.com/ion/winroom/bpp/internal/ratesvc"
	"github.com/ion/winroom/bpp/internal/selectsvc"
	"github.com/ion/winroom/bpp/internal/statussvc"
	"github.com/ion/winroom/bpp/internal/supportsvc"
	"github.com/ion/winroom/bpp/internal/updatesvc"
)

// RegisterRoutes wires all inbound BPP endpoints to the Alya service.
//
// A BPP only HOSTS the action endpoints — BAPs call these.
// The on_* callbacks (on_select, on_confirm, etc.) are outbound HTTP calls
// the BPP makes to the BAP's callback URL; they are NOT endpoints we expose.
//
// Inbound request flow (BAP → BPP):
//
//	/select    → BPP prices the quote, calls BAP's /on_select
//	/init      → BPP drafts the contract, calls BAP's /on_init
//	/confirm   → BPP confirms the contract, calls BAP's /on_confirm
//	/status    → BPP returns contract state, calls BAP's /on_status
//	/track     → BPP returns tracking handle, calls BAP's /on_track
//	/update    → BPP mutates the contract, calls BAP's /on_update
//	/cancel    → BPP cancels the contract, calls BAP's /on_cancel
//	/rate      → BPP records the rating, calls BAP's /on_rate
//	/support   → BPP creates a support ticket, calls BAP's /on_support
//
// Provider-facing:
//
//	POST /v1/catalog/publish → Provider pushes catalog; BPP saves to DB and forwards to CDS
//
// Inbound CDS → BPP callback:
//
//	/beckn/catalog/on_publish → CDS notifies BPP of catalog indexing result
func RegisterRoutes(svc *service.Service, cfg *config.Config) {
	pool := svc.Database.(*pgxpool.Pool)

	// Health check — outside Beckn namespace, used by load balancers / k8s probes
	svc.RegisterRoute(http.MethodGet, "/health", handleHealth)

	// Provider-facing catalog publish API (simplified — no Beckn context required from provider)
	v1 := svc.Router.Group("/api/v1")
	registerProviderCatalogRoutes(svc, v1, pool, cfg)
	registerDashboardRoutes(v1, pool, cfg)

	webhook := svc.Router.Group("/api/webhook")
	registerTransactionRoutes(svc, webhook, pool, cfg)
	registerFulfillmentRoutes(svc, webhook, pool, cfg)
	registerPostFulfillmentRoutes(svc, webhook, pool, cfg)
	registerBecknCatalogRoutes(svc, webhook)
}

func handleHealth(c *gin.Context, _ *service.Service) {
	c.JSON(http.StatusOK, wscutils.NewSuccessResponse(gin.H{"status": "ok"}))
}

// ---------------------------------------------------------------------------
// Transaction — select / init / confirm
// ---------------------------------------------------------------------------

func registerTransactionRoutes(svc *service.Service, g *gin.RouterGroup, pool *pgxpool.Pool, cfg *config.Config) {
	selectH := selectsvc.NewSelectHandler(pool, cfg, svc.LogHarbour)
	// /webhook/select — raw Gin handler; returns Beckn ACK immediately, processes async.
	g.POST("/select", selectH.Handle)

	// /webhook/init — persists contract state, fires on_init callback.
	initH := initsvc.NewInitHandler(pool, cfg, svc.LogHarbour)
	g.POST("/init", initH.Handle)

	// /webhook/confirm — transitions DRAFT→ACTIVE, fires on_confirm callback.
	confirmH := confirmsvc.NewConfirmHandler(pool, cfg, svc.LogHarbour)
	g.POST("/confirm", confirmH.Handle)
}

// ---------------------------------------------------------------------------
// Fulfillment — status / track / update / cancel
// ---------------------------------------------------------------------------

func registerFulfillmentRoutes(svc *service.Service, g *gin.RouterGroup, pool *pgxpool.Pool, cfg *config.Config) {
	statusH := statussvc.NewHandler(pool, cfg, svc.LogHarbour)
	g.POST("/status", statusH.Handle)

	updateH := updatesvc.NewHandler(pool, cfg, svc.LogHarbour)
	g.POST("/update", updateH.Handle)

	cancelH := cancelsvc.NewHandler(pool, cfg, svc.LogHarbour)
	g.POST("/cancel", cancelH.Handle)

	svc.RegisterRouteWithGroup(g, http.MethodPost, "/track", handleTrack)
}

func handleTrack(c *gin.Context, svc *service.Service) {
	c.JSON(http.StatusOK, wscutils.NewSuccessResponse(nil))
}

// ---------------------------------------------------------------------------
// Post-fulfillment — rate / support
// ---------------------------------------------------------------------------

func registerPostFulfillmentRoutes(svc *service.Service, g *gin.RouterGroup, pool *pgxpool.Pool, cfg *config.Config) {
	rateH := ratesvc.NewHandler(pool, cfg, svc.LogHarbour)
	g.POST("/rate", rateH.Handle)

	supportH := supportsvc.NewHandler(pool, cfg, svc.LogHarbour)
	g.POST("/support", supportH.Handle)
}

// ---------------------------------------------------------------------------
// Provider-facing catalog API  (POST /v1/catalog/publish)
//
// Providers call this to publish their catalog.
// The BPP saves the data to the database, then forwards a Beckn-compatible
// request to the CDS.  The provider never needs to know about Beckn context.
// ---------------------------------------------------------------------------

func registerProviderCatalogRoutes(svc *service.Service, g *gin.RouterGroup, pool *pgxpool.Pool, cfg *config.Config) {
	h := catalog.NewPublishHandler(pool, cfg, svc.LogHarbour)
	cat := g.Group("/catalog")
	svc.RegisterRouteWithGroup(cat, http.MethodPost, "/publish", h.Handle)
}

// ---------------------------------------------------------------------------
// Beckn catalog callbacks  (POST /beckn/catalog/on_publish)
//
// The CDS calls /beckn/catalog/on_publish to notify the BPP that a previously
// submitted catalog has been indexed (or rejected).
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// Dashboard API  (GET /api/v1/dashboard/*, /api/v1/orders/*, etc.)
// ---------------------------------------------------------------------------

func registerDashboardRoutes(v1 *gin.RouterGroup, pool *pgxpool.Pool, cfg *config.Config) {
	h := dashboardsvc.NewHandler(pool, cfg)

	v1.GET("/dashboard/stats", h.HandleStats)

	v1.GET("/orders", h.HandleListOrders)
	v1.GET("/orders/:id", h.HandleGetOrder)

	inv := v1.Group("/inventory")
	inv.GET("/resources", h.HandleListResources)
	inv.GET("/offers", h.HandleListOffers)
	inv.GET("/stock", h.HandleListStock)

	v1.GET("/messages", h.HandleListMessages)
	v1.GET("/support-tickets", h.HandleListSupportTickets)
	v1.GET("/ratings", h.HandleListRatings)
}

func registerBecknCatalogRoutes(svc *service.Service, g *gin.RouterGroup) {
	cat := g.Group("/catalog")
	svc.RegisterRouteWithGroup(cat, http.MethodPost, "/on_publish", handleOnPublish)
}

func handleOnPublish(c *gin.Context, svc *service.Service) {
	// TODO: parse indexing result, update catalog status in DB, notify provider
	c.JSON(http.StatusOK, wscutils.NewSuccessResponse(nil))
}
