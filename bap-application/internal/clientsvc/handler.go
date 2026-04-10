package clientsvc

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/remiges-tech/logharbour/logharbour"

	"github.com/ion/winroom/bap/internal/config"
)

type ClientHandler struct {
	svc *ClientService
	lh  *logharbour.Logger
}

func NewClientHandler(pool *pgxpool.Pool, cfg *config.Config, lh *logharbour.Logger) *ClientHandler {
	return &ClientHandler{
		svc: NewClientService(pool, cfg, lh),
		lh:  lh,
	}
}

// HandleSelect triggers a Beckn select request.
func (h *ClientHandler) HandleSelect(c *gin.Context) {
	var req ClientSelectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	txnID, err := h.svc.Select(c.Request.Context(), &req)
	if err != nil {
		h.lh.WithModule("client_handler").Err().Error(err).Log("select failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transaction_id": txnID})
}

// HandleInit triggers a Beckn init request.
func (h *ClientHandler) HandleInit(c *gin.Context) {
	var req ClientInitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.Init(c.Request.Context(), &req); err != nil {
		h.lh.WithModule("client_handler").Err().Error(err).Log("init failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "INIT_SENT"})
}

// HandleConfirm triggers a Beckn confirm request.
func (h *ClientHandler) HandleConfirm(c *gin.Context) {
	var req ClientConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.Confirm(c.Request.Context(), &req); err != nil {
		h.lh.WithModule("client_handler").Err().Error(err).Log("confirm failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "CONFIRM_SENT"})
}

// HandleStatus returns the transaction lifecycle state.
func (h *ClientHandler) HandleStatus(c *gin.Context) {
	txnID := c.Param("id")
	if txnID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id parameter required"})
		return
	}

	resp, err := h.svc.GetStatus(c.Request.Context(), txnID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// HandleDiscover triggers a Discover request to the CDS.
// GET /api/v1/discover?q=<text search term>
// The CDS requires a non-empty textSearch — defaults to "coffee" for the initial catalog load.
func (h *ClientHandler) HandleDiscover(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		q = "coffee"
	}
	req := ClientDiscoverRequest{TextSearch: q}

	resp, err := h.svc.Discover(c.Request.Context(), &req)
	if err != nil {
		h.lh.WithModule("client_handler").Err().Error(err).Log("discover failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/json", resp)
}

// HandleCatalog returns the food retail catalog.
func (h *ClientHandler) HandleCatalog(c *gin.Context) {
	catalog, err := h.svc.GetCatalog(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Data(http.StatusOK, "application/json", catalog)
}
