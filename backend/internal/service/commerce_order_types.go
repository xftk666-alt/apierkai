package service

import (
	"context"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

var (
	ErrCommerceOrdersDisabled       = infraerrors.Forbidden("COMMERCE_ORDERS_DISABLED", "commerce orders are not enabled")
	ErrCommerceWalletDisabled       = infraerrors.Forbidden("COMMERCE_WALLET_DISABLED", "commerce wallet is not enabled")
	ErrCommerceOrderNotFound        = infraerrors.NotFound("COMMERCE_ORDER_NOT_FOUND", "commerce order not found")
	ErrCommercePriceNotFound        = infraerrors.NotFound("COMMERCE_PRICE_NOT_FOUND", "commerce price not found")
	ErrCommerceOrderAlreadyHandled  = infraerrors.Conflict("COMMERCE_ORDER_ALREADY_HANDLED", "commerce order has already been handled")
	ErrCommerceOrderSnapshotInvalid = infraerrors.BadRequest("COMMERCE_ORDER_SNAPSHOT_INVALID", "commerce order snapshot is invalid")
	ErrCommerceProductInactive      = infraerrors.BadRequest("COMMERCE_PRODUCT_INACTIVE", "commerce product is not active")
	ErrCommercePriceDisabled        = infraerrors.BadRequest("COMMERCE_PRICE_DISABLED", "commerce price is not enabled")
	ErrCommerceOrderNoRequired      = infraerrors.BadRequest("COMMERCE_ORDER_NO_REQUIRED", "commerce order number is required")
	ErrCommercePaymentProviderNotFound = infraerrors.BadRequest("COMMERCE_PAYMENT_PROVIDER_NOT_FOUND", "commerce payment provider not found")
)

const (
	CommercePaymentProviderManual      = "manual"
	CommerceWalletReasonTypeOrderTopUp = "commerce_order_topup"
	commerceOrderNotePrefix            = "commerce order "
)

type CommerceOrderSnapshotProduct struct {
	ID          int64          `json:"id"`
	Code        string         `json:"code"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	ProductType string         `json:"product_type"`
	Status      string         `json:"status"`
	CoverImage  string         `json:"cover_image"`
	Tags        []string       `json:"tags"`
	SortOrder   int            `json:"sort_order"`
	Recommended bool           `json:"recommended"`
	Metadata    map[string]any `json:"metadata"`
}

type CommerceOrderSnapshot struct {
	Product CommerceOrderSnapshotProduct `json:"product"`
	Price   CommerceCatalogPrice         `json:"price"`
	Grants  []CommerceCatalogGrant       `json:"grants"`
}

type CommerceOrderCreateInput struct {
	ProductID       int64  `json:"product_id"`
	PriceID         int64  `json:"price_id"`
	PaymentProvider string `json:"payment_provider"`
}

type CommerceOrderManualCompleteInput struct {
	PaymentProvider string   `json:"payment_provider"`
	ProviderTradeNo string   `json:"provider_trade_no"`
	RequestPayload  string   `json:"request_payload"`
	CallbackPayload string   `json:"callback_payload"`
	PaidAmount      *float64 `json:"paid_amount"`
	PaidCurrency    string   `json:"paid_currency"`
}

type CommerceOrderPaymentCallbackInput struct {
	OrderNo         string   `json:"order_no"`
	ProviderTradeNo string   `json:"provider_trade_no"`
	RequestPayload  string   `json:"request_payload"`
	CallbackPayload string   `json:"callback_payload"`
	PaidAmount      *float64 `json:"paid_amount"`
	PaidCurrency    string   `json:"paid_currency"`
}

type CommerceOrderListFilters struct {
	UserID        *int64
	Status        string
	PaymentStatus string
	OrderNo       string
}

type CommerceWalletLedgerListFilters struct {
	UserID     *int64
	OrderID    *int64
	Direction  string
	ReasonType string
}

type CommerceOrderCompleteRecord struct {
	Status          string
	PaymentStatus   string
	PaymentProvider string
	Amount          float64
	Currency        string
	PaidAt          *time.Time
	CompletedAt     *time.Time
}

type CommerceWalletLedgerRecord struct {
	CommerceWalletLedger
	OrderNo string
}

type CommerceOrderPaymentTransactionView struct {
	ID              int64     `json:"id"`
	OrderID         int64     `json:"order_id"`
	Provider        string    `json:"provider"`
	ProviderTradeNo string    `json:"provider_trade_no"`
	Status          string    `json:"status"`
	RequestPayload  string    `json:"request_payload"`
	CallbackPayload string    `json:"callback_payload"`
	PaidAmount      float64   `json:"paid_amount"`
	PaidCurrency    string    `json:"paid_currency"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CommerceOrderView struct {
	ID                 int64                             `json:"id"`
	OrderNo            string                            `json:"order_no"`
	UserID             int64                             `json:"user_id"`
	ProductID          int64                             `json:"product_id"`
	PriceID            int64                             `json:"price_id"`
	Status             string                            `json:"status"`
	PaymentStatus      string                            `json:"payment_status"`
	PaymentProvider    string                            `json:"payment_provider"`
	Amount             float64                           `json:"amount"`
	Currency           string                            `json:"currency"`
	Snapshot           CommerceOrderSnapshot             `json:"snapshot"`
	PaymentAction      *CommerceOrderPaymentAction       `json:"payment_action,omitempty"`
	PaymentTransaction *CommerceOrderPaymentTransactionView `json:"payment_transaction,omitempty"`
	PaidAt             *time.Time                        `json:"paid_at"`
	ExpiredAt          *time.Time                        `json:"expired_at"`
	CompletedAt        *time.Time                        `json:"completed_at"`
	CreatedAt          time.Time                         `json:"created_at"`
	UpdatedAt          time.Time                         `json:"updated_at"`
}

type CommerceWalletLedgerView struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	OrderID       *int64    `json:"order_id"`
	OrderNo       string    `json:"order_no"`
	Direction     string    `json:"direction"`
	ChangeAmount  float64   `json:"change_amount"`
	BalanceBefore float64   `json:"balance_before"`
	BalanceAfter  float64   `json:"balance_after"`
	ReasonType    string    `json:"reason_type"`
	ReasonDetail  string    `json:"reason_detail"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CommerceOrderRepository interface {
	CreateOrder(ctx context.Context, order *CommerceOrder) (*CommerceOrder, error)
	GetOrderByID(ctx context.Context, id int64) (*CommerceOrder, error)
	GetOrderByOrderNo(ctx context.Context, orderNo string) (*CommerceOrder, error)
	GetOrderByIDForUpdate(ctx context.Context, id int64) (*CommerceOrder, error)
	ListOrders(ctx context.Context, params pagination.PaginationParams, filters CommerceOrderListFilters) ([]CommerceOrder, *pagination.PaginationResult, error)
	ListLatestPaymentTransactionsByOrderIDs(ctx context.Context, orderIDs []int64) (map[int64]CommercePaymentTransaction, error)
	CompleteOrder(ctx context.Context, id int64, input *CommerceOrderCompleteRecord) (*CommerceOrder, error)
	CreatePaymentTransaction(ctx context.Context, transaction *CommercePaymentTransaction) error
	CreateWalletLedger(ctx context.Context, ledger *CommerceWalletLedger) error
	ListWalletLedgers(ctx context.Context, params pagination.PaginationParams, filters CommerceWalletLedgerListFilters) ([]CommerceWalletLedgerRecord, *pagination.PaginationResult, error)
}

type CommerceOrderService struct {
	repo                 CommerceOrderRepository
	catalogService       *CommerceCatalogService
	userRepo             UserRepository
	subscriptionService  *SubscriptionService
	billingCacheService  *BillingCacheService
	authCacheInvalidator APIKeyAuthCacheInvalidator
	entClient            *dbent.Client
}
