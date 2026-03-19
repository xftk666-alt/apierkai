package admin

import (
	"context"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type CommerceHandler struct {
	catalogService *service.CommerceCatalogService
	orderService   *service.CommerceOrderService
}

func NewCommerceHandler(catalogService *service.CommerceCatalogService, orderService *service.CommerceOrderService) *CommerceHandler {
	return &CommerceHandler{
		catalogService: catalogService,
		orderService:   orderService,
	}
}

func (h *CommerceHandler) GetCatalog(c *gin.Context) {
	catalog, err := h.catalogService.GetAdminCatalog(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, catalog)
}

func (h *CommerceHandler) CreateProduct(c *gin.Context) {
	var input service.CommerceProductUpsertInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	product, err := h.catalogService.CreateProduct(c.Request.Context(), &input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, product)
}

func (h *CommerceHandler) UpdateProduct(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid product ID")
		return
	}

	var input service.CommerceProductUpsertInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	product, err := h.catalogService.UpdateProduct(c.Request.Context(), id, &input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, product)
}

func (h *CommerceHandler) CreateModel(c *gin.Context) {
	var input service.CommerceModelListingUpsertInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	model, err := h.catalogService.CreateModelListing(c.Request.Context(), &input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, model)
}

func (h *CommerceHandler) UpdateModel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid model ID")
		return
	}

	var input service.CommerceModelListingUpsertInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	model, err := h.catalogService.UpdateModelListing(c.Request.Context(), id, &input)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, model)
}

func (h *CommerceHandler) ListOrders(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)

	filters := service.CommerceOrderListFilters{
		UserID:        parseInt64Query(c, "user_id"),
		Status:        c.Query("status"),
		PaymentStatus: c.Query("payment_status"),
		OrderNo:       c.Query("order_no"),
	}

	items, pager, err := h.orderService.ListAdminOrders(c.Request.Context(), pagination.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}, filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.PaginatedWithResult(c, items, toResponsePagination(pager))
}

func (h *CommerceHandler) ListWalletLedger(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)

	filters := service.CommerceWalletLedgerListFilters{
		UserID:     parseInt64Query(c, "user_id"),
		OrderID:    parseInt64Query(c, "order_id"),
		Direction:  c.Query("direction"),
		ReasonType: c.Query("reason_type"),
	}

	items, pager, err := h.orderService.ListAdminWalletLedgers(c.Request.Context(), pagination.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}, filters)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.PaginatedWithResult(c, items, toResponsePagination(pager))
}

func (h *CommerceHandler) ManualCompleteOrder(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}

	var input service.CommerceOrderManualCompleteInput
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&input); err != nil {
			response.BadRequest(c, "Invalid request: "+err.Error())
			return
		}
	}

	idempotencyPayload := struct {
		OrderID int64                             `json:"order_id"`
		Body    service.CommerceOrderManualCompleteInput `json:"body"`
	}{
		OrderID: orderID,
		Body:    input,
	}
	executeAdminIdempotentJSON(c, "admin.commerce.orders.manual_complete", idempotencyPayload, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		return h.orderService.ManualCompleteOrder(ctx, orderID, getAdminIDFromContext(c), &input)
	})
}

func parseInt64Query(c *gin.Context, key string) *int64 {
	raw := c.Query(key)
	if raw == "" {
		return nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil
	}
	return &value
}
