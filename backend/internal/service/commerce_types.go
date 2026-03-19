package service

import "time"

const (
	CommerceProductTypeTopUpBalance     = "topup_balance"
	CommerceProductTypeSubscription     = "subscription_group"
	CommerceProductTypeStandardAccess   = "standard_group_access"
	CommerceProductTypeCombo            = "combo"
	CommerceProductStatusDraft          = "draft"
	CommerceProductStatusActive         = "active"
	CommerceProductStatusDisabled       = "disabled"
	CommercePriceTypeOneTime            = "one_time"
	CommercePriceTypeMonthly            = "monthly"
	CommercePriceTypeQuarterly          = "quarterly"
	CommercePriceTypeYearly             = "yearly"
	CommerceGrantTypeSubscription       = "subscription"
	CommerceGrantTypeAllowedGroup       = "allowed_group"
	CommerceOrderStatusPending          = "pending"
	CommerceOrderStatusPaid             = "paid"
	CommerceOrderStatusCompleted        = "completed"
	CommerceOrderStatusExpired          = "expired"
	CommerceOrderStatusCancelled        = "cancelled"
	CommercePaymentStatusPending        = "pending"
	CommercePaymentStatusSucceeded      = "succeeded"
	CommercePaymentStatusFailed         = "failed"
	CommercePaymentStatusRefunded       = "refunded"
	CommerceWalletDirectionCredit       = "credit"
	CommerceWalletDirectionDebit        = "debit"
	CommerceModelBindingTypePrimary     = "primary"
	CommerceModelBindingTypeUpsell      = "upsell"
	CommerceModelBindingTypeTopUp       = "topup"
)

type CommerceProduct struct {
	ID          int64
	Code        string
	Name        string
	Description string
	ProductType string
	Status      string
	CoverImage  string
	Tags        []string
	SortOrder   int
	Recommended bool
	Metadata    map[string]any
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CommerceProductPrice struct {
	ID             int64
	ProductID      int64
	PriceType      string
	Amount         float64
	Currency       string
	OriginalAmount *float64
	Enabled        bool
	SortOrder      int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CommerceProductGroupBinding struct {
	ID           int64
	ProductID    int64
	GroupID      int64
	GrantType    string
	ValidityDays int
	Priority     int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CommerceModelListing struct {
	ID          int64
	ModelKey    string
	DisplayName string
	Description string
	Vendor      string
	Icon        string
	Tags        []string
	Status      string
	SortOrder   int
	Recommended bool
	Metadata    map[string]any
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CommerceModelProductBinding struct {
	ID             int64
	ModelListingID int64
	ProductID      int64
	BindingType    string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CommerceOrder struct {
	ID              int64
	OrderNo         string
	UserID          int64
	ProductID       int64
	PriceID         int64
	Status          string
	PaymentStatus   string
	PaymentProvider string
	Amount          float64
	Currency        string
	Snapshot        map[string]any
	PaidAt          *time.Time
	ExpiredAt       *time.Time
	CompletedAt     *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type CommercePaymentTransaction struct {
	ID              int64
	OrderID         int64
	Provider        string
	ProviderTradeNo string
	Status          string
	RequestPayload  string
	CallbackPayload string
	PaidAmount      float64
	PaidCurrency    string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type CommerceWalletLedger struct {
	ID            int64
	UserID        int64
	OrderID       *int64
	Direction     string
	ChangeAmount  float64
	BalanceBefore float64
	BalanceAfter  float64
	ReasonType    string
	ReasonDetail  string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

