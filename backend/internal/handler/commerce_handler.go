package handler

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

const commerceCallbackSecretHeader = "X-Commerce-Callback-Secret"

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
	catalog, err := h.catalogService.GetUserCatalog(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, catalog)
}

func (h *CommerceHandler) CreateOrder(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	var input service.CommerceOrderCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	idempotencyPayload := struct {
		UserID int64                          `json:"user_id"`
		Body   service.CommerceOrderCreateInput `json:"body"`
	}{
		UserID: subject.UserID,
		Body:   input,
	}
	executeUserIdempotentJSON(c, "commerce.orders.create", idempotencyPayload, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		return h.orderService.CreateOrder(ctx, subject.UserID, &input)
	})
}

func (h *CommerceHandler) ListOrders(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	page, pageSize := response.ParsePagination(c)
	items, pager, err := h.orderService.ListUserOrders(c.Request.Context(), subject.UserID, pagination.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.PaginatedWithResult(c, items, toCommerceResponsePagination(pager))
}

func (h *CommerceHandler) ListWalletLedger(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	page, pageSize := response.ParsePagination(c)
	items, pager, err := h.orderService.ListUserWalletLedgers(c.Request.Context(), subject.UserID, pagination.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.PaginatedWithResult(c, items, toCommerceResponsePagination(pager))
}

type commercePaymentCallbackRequest struct {
	OrderNo         string  `json:"order_no"`
	ProviderTradeNo string  `json:"provider_trade_no"`
	PaidAmount      *float64 `json:"paid_amount"`
	PaidCurrency    string  `json:"paid_currency"`
	RequestPayload  any     `json:"request_payload"`
	CallbackPayload any     `json:"callback_payload"`
}

func (h *CommerceHandler) HandleProviderCallback(c *gin.Context) {
	if err := h.catalogService.ValidateCallbackSecret(c.Request.Context(), c.GetHeader(commerceCallbackSecretHeader)); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	provider := strings.TrimSpace(c.Param("provider"))
	if provider == "" {
		response.BadRequest(c, "Invalid provider")
		return
	}

	var req commercePaymentCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	requestPayload, err := normalizeCommerceCallbackPayload(req.RequestPayload)
	if err != nil {
		response.BadRequest(c, "Invalid request_payload: "+err.Error())
		return
	}
	callbackPayload, err := normalizeCommerceCallbackPayload(req.CallbackPayload)
	if err != nil {
		response.BadRequest(c, "Invalid callback_payload: "+err.Error())
		return
	}

	view, err := h.orderService.HandlePaymentCallback(c.Request.Context(), provider, &service.CommerceOrderPaymentCallbackInput{
		OrderNo:         req.OrderNo,
		ProviderTradeNo: req.ProviderTradeNo,
		RequestPayload:  requestPayload,
		CallbackPayload: callbackPayload,
		PaidAmount:      req.PaidAmount,
		PaidCurrency:    req.PaidCurrency,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, view)
}

func toCommerceResponsePagination(p *pagination.PaginationResult) *response.PaginationResult {
	if p == nil {
		return nil
	}
	return &response.PaginationResult{
		Total:    p.Total,
		Page:     p.Page,
		PageSize: p.PageSize,
		Pages:    p.Pages,
	}
}

func normalizeCommerceCallbackPayload(value any) (string, error) {
	if value == nil {
		return "", nil
	}

	if raw, ok := value.(string); ok {
		return strings.TrimSpace(raw), nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
