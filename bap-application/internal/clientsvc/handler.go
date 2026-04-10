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

// HandleRequestStatus triggers a Beckn status request (distinct from the local GET /status/:id poll).
func (h *ClientHandler) HandleRequestStatus(c *gin.Context) {
	var req ClientRequestStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.RequestStatus(c.Request.Context(), &req); err != nil {
		h.lh.WithModule("client_handler").Err().Error(err).Log("status request failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "STATUS_SENT"})
}

// HandleCancel triggers a Beckn cancel request.
func (h *ClientHandler) HandleCancel(c *gin.Context) {
	var req ClientCancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Cancel(c.Request.Context(), &req); err != nil {
		h.lh.WithModule("client_handler").Err().Error(err).Log("cancel failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "CANCEL_SENT"})
}

// HandleUpdate triggers a Beckn update request.
func (h *ClientHandler) HandleUpdate(c *gin.Context) {
	var req ClientUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Update(c.Request.Context(), &req); err != nil {
		h.lh.WithModule("client_handler").Err().Error(err).Log("update failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "UPDATE_SENT"})
}

// HandleRate submits ratings for a completed order.
func (h *ClientHandler) HandleRate(c *gin.Context) {
	var req ClientRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Rate(c.Request.Context(), &req); err != nil {
		h.lh.WithModule("client_handler").Err().Error(err).Log("rate failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "RATE_SENT"})
}

// HandleSupport raises a support request for an order.
func (h *ClientHandler) HandleSupport(c *gin.Context) {
	var req ClientSupportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.Support(c.Request.Context(), &req); err != nil {
		h.lh.WithModule("client_handler").Err().Error(err).Log("support failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "SUPPORT_SENT"})
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
