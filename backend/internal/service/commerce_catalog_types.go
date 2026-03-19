package service

type CommerceCatalog struct {
	Features CommerceCatalogFeatures  `json:"features"`
	Products []CommerceCatalogProduct `json:"products"`
	Models   []CommerceCatalogModel   `json:"models"`
}

type CommerceCatalogFeatures struct {
	PurchaseEnabled    bool   `json:"purchase_enabled"`
	PurchaseMode       string `json:"purchase_mode"`
	MarketplaceEnabled bool   `json:"marketplace_enabled"`
	WalletEnabled      bool   `json:"wallet_enabled"`
	OrdersEnabled      bool   `json:"orders_enabled"`
	LegacyPurchaseURL  string `json:"legacy_purchase_url"`
	PaymentProviders   []CommercePaymentProviderOption `json:"payment_providers"`
}

type CommerceCatalogProduct struct {
	ID          int64                  `json:"id"`
	Code        string                 `json:"code"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	ProductType string                 `json:"product_type"`
	Status      string                 `json:"status"`
	CoverImage  string                 `json:"cover_image"`
	Tags        []string               `json:"tags"`
	SortOrder   int                    `json:"sort_order"`
	Recommended bool                   `json:"recommended"`
	Metadata    map[string]any         `json:"metadata"`
	Prices      []CommerceCatalogPrice `json:"prices"`
	Grants      []CommerceCatalogGrant `json:"grants"`
}

type CommerceCatalogPrice struct {
	ID             int64    `json:"id"`
	PriceType      string   `json:"price_type"`
	Amount         float64  `json:"amount"`
	Currency       string   `json:"currency"`
	OriginalAmount *float64 `json:"original_amount"`
	Enabled        bool     `json:"enabled"`
	SortOrder      int      `json:"sort_order"`
}

type CommerceCatalogGrant struct {
	ID           int64  `json:"id"`
	GroupID      int64  `json:"group_id"`
	GroupName    string `json:"group_name"`
	GrantType    string `json:"grant_type"`
	ValidityDays int    `json:"validity_days"`
	Priority     int    `json:"priority"`
}

type CommerceCatalogModel struct {
	ID          int64                 `json:"id"`
	ModelKey    string                `json:"model_key"`
	DisplayName string                `json:"display_name"`
	Description string                `json:"description"`
	Vendor      string                `json:"vendor"`
	Icon        string                `json:"icon"`
	Tags        []string              `json:"tags"`
	Status      string                `json:"status"`
	SortOrder   int                   `json:"sort_order"`
	Recommended bool                  `json:"recommended"`
	Metadata    map[string]any        `json:"metadata"`
	Offers      []CommerceCatalogOffer `json:"offers"`
}

type CommerceCatalogOffer struct {
	BindingID   int64                  `json:"binding_id"`
	BindingType string                 `json:"binding_type"`
	Product     CommerceCatalogProduct `json:"product"`
}

type CommerceProductPriceInput struct {
	PriceType      string   `json:"price_type"`
	Amount         float64  `json:"amount"`
	Currency       string   `json:"currency"`
	OriginalAmount *float64 `json:"original_amount"`
	Enabled        *bool    `json:"enabled"`
	SortOrder      int      `json:"sort_order"`
}

type CommerceProductGrantInput struct {
	GroupID      int64  `json:"group_id"`
	GrantType    string `json:"grant_type"`
	ValidityDays int    `json:"validity_days"`
	Priority     int    `json:"priority"`
}

type CommerceProductUpsertInput struct {
	Code          string                      `json:"code"`
	Name          string                      `json:"name"`
	Description   string                      `json:"description"`
	ProductType   string                      `json:"product_type"`
	Status        string                      `json:"status"`
	CoverImage    string                      `json:"cover_image"`
	Tags          []string                    `json:"tags"`
	SortOrder     int                         `json:"sort_order"`
	Recommended   bool                        `json:"recommended"`
	Metadata      map[string]any              `json:"metadata"`
	Prices        []CommerceProductPriceInput `json:"prices"`
	GroupBindings []CommerceProductGrantInput `json:"group_bindings"`
}

type CommerceModelProductBindingInput struct {
	ProductID   int64  `json:"product_id"`
	BindingType string `json:"binding_type"`
}

type CommerceModelListingUpsertInput struct {
	ModelKey        string                             `json:"model_key"`
	DisplayName     string                             `json:"display_name"`
	Description     string                             `json:"description"`
	Vendor          string                             `json:"vendor"`
	Icon            string                             `json:"icon"`
	Tags            []string                           `json:"tags"`
	Status          string                             `json:"status"`
	SortOrder       int                                `json:"sort_order"`
	Recommended     bool                               `json:"recommended"`
	Metadata        map[string]any                     `json:"metadata"`
	ProductBindings []CommerceModelProductBindingInput `json:"product_bindings"`
}
