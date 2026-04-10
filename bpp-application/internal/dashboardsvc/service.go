package dashboardsvc

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ion/winroom/bpp/db/sqlc"
	"github.com/ion/winroom/bpp/internal/config"
)

const defaultPageSize = 20

// Handler bundles all dashboard API handlers and their shared dependencies.
type Handler struct {
	pool *pgxpool.Pool
	cfg  *config.Config
}

func NewHandler(pool *pgxpool.Pool, cfg *config.Config) *Handler {
	return &Handler{pool: pool, cfg: cfg}
}

// ---------------------------------------------------------------------------
// GET /api/v1/dashboard/stats
// ---------------------------------------------------------------------------

// HandleStats returns key metrics: active/pending/today orders, resource/offer counts,
// and message funnel counts (select → init → confirm).
func (h *Handler) HandleStats(c *gin.Context) {
	q := sqlc.New(h.pool)
	ctx := c.Request.Context()
	bppID := h.cfg.BppID

	orderStats, err := q.GetDashboardStats(ctx, bppID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	invStats, err := q.GetInventoryStats(ctx, bppID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	funnelRows, err := q.GetMessageFunnel(ctx, &bppID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	funnel := map[string]int64{"select": 0, "init": 0, "confirm": 0}
	for _, row := range funnelRows {
		funnel[string(row.Action)] = row.MessageCount
	}

	c.JSON(http.StatusOK, gin.H{
		"active_orders":  orderStats.ActiveOrders,
		"pending_orders": orderStats.PendingOrders,
		"today_orders":   orderStats.TodayOrders,
		"resource_count": invStats.ResourceCount,
		"offer_count":    invStats.OfferCount,
		"funnel":         funnel,
	})
}

// ---------------------------------------------------------------------------
// GET /api/v1/orders
// ---------------------------------------------------------------------------

// HandleListOrders returns a paginated list of contracts.
func (h *Handler) HandleListOrders(c *gin.Context) {
	page, limit := parsePagination(c)
	q := sqlc.New(h.pool)
	ctx := c.Request.Context()

	total, err := q.CountContracts(ctx, h.cfg.BppID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, err := q.ListContracts(ctx, sqlc.ListContractsParams{
		BppID:  h.cfg.BppID,
		Limit:  int32(limit),
		Offset: int32((page - 1) * limit),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"page":  page,
		"limit": limit,
		"items": rows,
	})
}

// ---------------------------------------------------------------------------
// GET /api/v1/orders/:id
// ---------------------------------------------------------------------------

// HandleGetOrder returns the full detail for a single contract, including
// commitments and considerations.
func (h *Handler) HandleGetOrder(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid contract id"})
		return
	}

	q := sqlc.New(h.pool)
	ctx := c.Request.Context()

	contract, err := q.GetContractDetail(ctx, sqlc.GetContractDetailParams{ID: id, BppID: h.cfg.BppID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	commitments, err := q.ListContractCommitments(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	considerations, err := q.ListContractConsiderations(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"contract":       contract,
		"commitments":    commitments,
		"considerations": considerations,
	})
}

// ---------------------------------------------------------------------------
// GET /api/v1/inventory/resources
// ---------------------------------------------------------------------------

// HandleListResources returns a paginated list of resources in the BPP's inventory.
func (h *Handler) HandleListResources(c *gin.Context) {
	page, limit := parsePagination(c)
	q := sqlc.New(h.pool)
	ctx := c.Request.Context()

	total, err := q.CountResources(ctx, h.cfg.BppID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, err := q.ListResources(ctx, sqlc.ListResourcesParams{
		BppID:  h.cfg.BppID,
		Limit:  int32(limit),
		Offset: int32((page - 1) * limit),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"page":  page,
		"limit": limit,
		"items": rows,
	})
}

// ---------------------------------------------------------------------------
// GET /api/v1/inventory/offers
// ---------------------------------------------------------------------------

// HandleListOffers returns a paginated list of offers with their considerations.
func (h *Handler) HandleListOffers(c *gin.Context) {
	page, limit := parsePagination(c)
	q := sqlc.New(h.pool)
	ctx := c.Request.Context()

	total, err := q.CountOffers(ctx, h.cfg.BppID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, err := q.ListOffers(ctx, sqlc.ListOffersParams{
		BppID:  h.cfg.BppID,
		Limit:  int32(limit),
		Offset: int32((page - 1) * limit),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"page":  page,
		"limit": limit,
		"items": rows,
	})
}

// ---------------------------------------------------------------------------
// GET /api/v1/messages
// ---------------------------------------------------------------------------

// HandleListMessages returns recent Beckn protocol messages from the audit log.
func (h *Handler) HandleListMessages(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	bppID := h.cfg.BppID
	q := sqlc.New(h.pool)
	rows, err := q.ListRecentMessages(c.Request.Context(), sqlc.ListRecentMessagesParams{
		BppID: &bppID,
		Limit: int32(limit),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": rows})
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func parsePagination(c *gin.Context) (page, limit int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ = strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(defaultPageSize)))
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = defaultPageSize
	}
	return
}
