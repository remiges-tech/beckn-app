package dashboardsvc

import (
	"encoding/json"
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

	// Stock summary — count out-of-stock and low-stock resources.
	stockRows, err := q.ListResourceStock(ctx, bppID)
	if err != nil {
		// Non-fatal: stock table may not exist yet on older deployments.
		stockRows = nil
	}
	var outOfStock, lowStock int
	for _, s := range stockRows {
		if s.Quantity == 0 {
			outOfStock++
		} else if s.Quantity <= 5 {
			lowStock++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"active_orders":  orderStats.ActiveOrders,
		"pending_orders": orderStats.PendingOrders,
		"today_orders":   orderStats.TodayOrders,
		"resource_count": invStats.ResourceCount,
		"offer_count":    invStats.OfferCount,
		"funnel":         funnel,
		"out_of_stock":   outOfStock,
		"low_stock":      lowStock,
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
// GET /api/v1/inventory/stock
// ---------------------------------------------------------------------------

// HandleListStock returns the current inventory stock levels for all resources
// belonging to this BPP, joined with their descriptor names.
func (h *Handler) HandleListStock(c *gin.Context) {
	q := sqlc.New(h.pool)
	rows, err := q.ListResourceStock(c.Request.Context(), h.cfg.BppID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type stockItem struct {
		ResourceID string  `json:"resourceId"`
		Name       string  `json:"name"`
		CatalogID  string  `json:"catalogId"`
		Quantity   int32   `json:"quantity"`
		Sold       int32   `json:"sold"`
		Available  int32   `json:"available"`
		UpdatedAt  string  `json:"updatedAt"`
	}

	items := make([]stockItem, 0, len(rows))
	for _, r := range rows {
		name := r.ResourceID
		if r.DescriptorName != nil && *r.DescriptorName != "" {
			name = *r.DescriptorName
		}
		catID := ""
		if r.CatalogID != nil {
			catID = *r.CatalogID
		}
		ts := ""
		if r.UpdatedAt.Valid {
			ts = r.UpdatedAt.Time.UTC().Format("2006-01-02T15:04:05Z")
		}
		items = append(items, stockItem{
			ResourceID: r.ResourceID,
			Name:       name,
			CatalogID:  catID,
			Quantity:   r.Quantity,
			Sold:       r.Sold,
			Available:  r.Quantity,
			UpdatedAt:  ts,
		})
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

// ---------------------------------------------------------------------------
// GET /api/v1/support-tickets
// ---------------------------------------------------------------------------

func (h *Handler) HandleListSupportTickets(c *gin.Context) {
	page, limit := parsePagination(c)
	q := sqlc.New(h.pool)
	ctx := c.Request.Context()

	total, err := q.CountSupportTickets(ctx, h.cfg.BppID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, err := q.ListSupportTickets(ctx, sqlc.ListSupportTicketsParams{
		BppID:  h.cfg.BppID,
		Limit:  int32(limit),
		Offset: int32((page - 1) * limit),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type ticketItem struct {
		ID          string `json:"id"`
		ContractID  string `json:"contractId"`
		TxnID       string `json:"transactionId"`
		BapID       string `json:"bapId"`
		Name        string `json:"name"`
		ShortDesc   string `json:"shortDesc"`
		Status      string `json:"status"`
		IsPreview   bool   `json:"isPreview"`
		CreatedAt   string `json:"createdAt"`
		ResolvedAt  string `json:"resolvedAt,omitempty"`
	}

	items := make([]ticketItem, 0, len(rows))
	for _, r := range rows {
		cid := ""
		if r.ContractID.Valid {
			cid = uuid.UUID(r.ContractID.Bytes).String()
		}
		txid := ""
		if r.TransactionID.Valid {
			txid = uuid.UUID(r.TransactionID.Bytes).String()
		}
		name := ""
		if r.DescriptorName != nil {
			name = *r.DescriptorName
		}
		sd := ""
		if r.DescriptorShortDesc != nil {
			sd = *r.DescriptorShortDesc
		}
		ca := ""
		if r.CreatedAt.Valid {
			ca = r.CreatedAt.Time.UTC().Format("2006-01-02T15:04:05Z")
		}
		ra := ""
		if r.ResolvedAt.Valid {
			ra = r.ResolvedAt.Time.UTC().Format("2006-01-02T15:04:05Z")
		}
		items = append(items, ticketItem{
			ID:         r.ID.String(),
			ContractID: cid,
			TxnID:      txid,
			BapID:      r.BapID,
			Name:       name,
			ShortDesc:  sd,
			Status:     string(r.Status),
			IsPreview:  r.IsPreview,
			CreatedAt:  ca,
			ResolvedAt: ra,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// ---------------------------------------------------------------------------
// GET /api/v1/ratings
// ---------------------------------------------------------------------------

func (h *Handler) HandleListRatings(c *gin.Context) {
	page, limit := parsePagination(c)
	q := sqlc.New(h.pool)
	ctx := c.Request.Context()

	total, err := q.CountRatings(ctx, h.cfg.BppID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, err := q.ListRatings(ctx, sqlc.ListRatingsParams{
		BppID:  h.cfg.BppID,
		Limit:  int32(limit),
		Offset: int32((page - 1) * limit),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type ratingItem struct {
		ID        string          `json:"id"`
		ContractID string         `json:"contractId"`
		TxnID     string          `json:"transactionId"`
		BapID     string          `json:"bapId"`
		TargetID  string          `json:"targetId"`
		Range     json.RawMessage `json:"range"`
		IsPreview bool            `json:"isPreview"`
		CreatedAt string          `json:"createdAt"`
	}

	items := make([]ratingItem, 0, len(rows))
	for _, r := range rows {
		ca := ""
		if r.CreatedAt.Valid {
			ca = r.CreatedAt.Time.UTC().Format("2006-01-02T15:04:05Z")
		}
		items = append(items, ratingItem{
			ID:         r.ID.String(),
			ContractID: r.ContractID.String(),
			TxnID:      r.TransactionID.String(),
			BapID:      r.BapID,
			TargetID:   r.TargetID,
			Range:      r.Range,
			IsPreview:  r.IsPreview,
			CreatedAt:  ca,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": total,
		"page":  page,
		"limit": limit,
	})
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
