package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/remiges-tech/alya/service"
	"github.com/remiges-tech/alya/wscutils"

	"github.com/ion/winroom/bpp/internal/catalog"
	"github.com/ion/winroom/bpp/internal/config"
	"github.com/ion/winroom/bpp/internal/selectsvc"
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
	v1 := svc.Router.Group("/v1")
	registerProviderCatalogRoutes(svc, v1, pool, cfg)

	beckn := svc.Router.Group("/beckn")
	registerTransactionRoutes(svc, beckn, pool, cfg)
	registerFulfillmentRoutes(svc, beckn)
	registerPostFulfillmentRoutes(svc, beckn)
	registerBecknCatalogRoutes(svc, beckn)
}

func handleHealth(c *gin.Context, _ *service.Service) {
	c.JSON(http.StatusOK, wscutils.NewSuccessResponse(gin.H{"status": "ok"}))
}

// ---------------------------------------------------------------------------
// Transaction — select / init / confirm
// ---------------------------------------------------------------------------

func registerTransactionRoutes(svc *service.Service, g *gin.RouterGroup, pool *pgxpool.Pool, cfg *config.Config) {
	selectH := selectsvc.NewSelectHandler(pool, cfg, svc.LogHarbour)
	// /beckn/select must bypass Alya's handler wrapper to return a raw Beckn ACK
	// (not wrapped in wscutils.Response) and kick off async processing.
	g.POST("/select", selectH.Handle)

	svc.RegisterRouteWithGroup(g, http.MethodPost, "/init", handleInit)
	svc.RegisterRouteWithGroup(g, http.MethodPost, "/confirm", handleConfirm)
}

func handleInit(c *gin.Context, svc *service.Service) {
	// TODO: draft contract with billing/fulfillment details, fire async on_init callback to BAP
	c.JSON(http.StatusOK, wscutils.NewSuccessResponse(nil))
}

func handleConfirm(c *gin.Context, svc *service.Service) {
	// TODO: confirm contract, fire async on_confirm callback to BAP
	c.JSON(http.StatusOK, wscutils.NewSuccessResponse(nil))
}

// ---------------------------------------------------------------------------
// Fulfillment — status / track / update / cancel
// ---------------------------------------------------------------------------

func registerFulfillmentRoutes(svc *service.Service, g *gin.RouterGroup) {
	svc.RegisterRouteWithGroup(g, http.MethodPost, "/status", handleStatus)
	svc.RegisterRouteWithGroup(g, http.MethodPost, "/track", handleTrack)
	svc.RegisterRouteWithGroup(g, http.MethodPost, "/update", handleUpdate)
	svc.RegisterRouteWithGroup(g, http.MethodPost, "/cancel", handleCancel)
}

func handleStatus(c *gin.Context, svc *service.Service) {
	// TODO: fetch contract state, fire async on_status callback to BAP
	c.JSON(http.StatusOK, wscutils.NewSuccessResponse(nil))
}

func handleTrack(c *gin.Context, svc *service.Service) {
	// TODO: return tracking URL/handle, fire async on_track callback to BAP
	c.JSON(http.StatusOK, wscutils.NewSuccessResponse(nil))
}

func handleUpdate(c *gin.Context, svc *service.Service) {
	// TODO: mutate contract, fire async on_update callback to BAP
	c.JSON(http.StatusOK, wscutils.NewSuccessResponse(nil))
}

func handleCancel(c *gin.Context, svc *service.Service) {
	// TODO: cancel contract, fire async on_cancel callback to BAP
	c.JSON(http.StatusOK, wscutils.NewSuccessResponse(nil))
}

// ---------------------------------------------------------------------------
// Post-fulfillment — rate / support
// ---------------------------------------------------------------------------

func registerPostFulfillmentRoutes(svc *service.Service, g *gin.RouterGroup) {
	svc.RegisterRouteWithGroup(g, http.MethodPost, "/rate", handleRate)
	svc.RegisterRouteWithGroup(g, http.MethodPost, "/support", handleSupport)
}

func handleRate(c *gin.Context, svc *service.Service) {
	// TODO: record rating, fire async on_rate callback to BAP
	c.JSON(http.StatusOK, wscutils.NewSuccessResponse(nil))
}

func handleSupport(c *gin.Context, svc *service.Service) {
	// TODO: create support ticket, fire async on_support callback to BAP
	c.JSON(http.StatusOK, wscutils.NewSuccessResponse(nil))
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

func registerBecknCatalogRoutes(svc *service.Service, g *gin.RouterGroup) {
	cat := g.Group("/catalog")
	svc.RegisterRouteWithGroup(cat, http.MethodPost, "/on_publish", handleOnPublish)
}

func handleOnPublish(c *gin.Context, svc *service.Service) {
	// TODO: parse indexing result, update catalog status in DB, notify provider
	c.JSON(http.StatusOK, wscutils.NewSuccessResponse(nil))
}
