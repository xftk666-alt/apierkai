package service

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

const (
	CommercePurchaseModeIframe = "iframe"
	CommercePurchaseModeNative = "native"
)

var commercePaymentProviderCodePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,31}$`)

type CommercePaymentProviderSetting struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	CheckoutURL string `json:"checkout_url"`
	Enabled     bool   `json:"enabled"`
	SortOrder   int    `json:"sort_order"`
}

type CommercePaymentProviderOption struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	SortOrder   int    `json:"sort_order"`
}

type CommerceOrderPaymentAction struct {
	Provider     string `json:"provider"`
	ProviderName string `json:"provider_name"`
	CheckoutURL  string `json:"checkout_url"`
}

type CommerceCheckoutURLTemplateInput struct {
	OrderID         int64
	OrderNo         string
	UserID          int64
	ProductID       int64
	PriceID         int64
	Amount          float64
	Currency        string
	PaymentProvider string
}

func NormalizeCommercePurchaseMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case CommercePurchaseModeNative:
		return CommercePurchaseModeNative
	default:
		return CommercePurchaseModeIframe
	}
}

func ParseCommercePaymentProviderSettings(raw string) []CommercePaymentProviderSetting {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return []CommercePaymentProviderSetting{}
	}

	var items []CommercePaymentProviderSetting
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return []CommercePaymentProviderSetting{}
	}

	normalized, err := NormalizeCommercePaymentProviderSettings(items)
	if err != nil {
		return []CommercePaymentProviderSetting{}
	}
	return normalized
}

func NormalizeCommercePaymentProviderSettings(items []CommercePaymentProviderSetting) ([]CommercePaymentProviderSetting, error) {
	if len(items) == 0 {
		return []CommercePaymentProviderSetting{}, nil
	}

	normalized := make([]CommercePaymentProviderSetting, 0, len(items))
	seen := make(map[string]struct{}, len(items))

	for _, item := range items {
		code := strings.ToLower(strings.TrimSpace(item.Code))
		name := strings.TrimSpace(item.Name)
		description := strings.TrimSpace(item.Description)
		icon := strings.TrimSpace(item.Icon)
		checkoutURL := strings.TrimSpace(item.CheckoutURL)
		sortOrder := item.SortOrder
		if sortOrder == 0 {
			sortOrder = 100
		}

		if code == "" && name == "" && description == "" && icon == "" && checkoutURL == "" {
			continue
		}
		if code == "" {
			return nil, fmt.Errorf("payment provider code is required")
		}
		if !commercePaymentProviderCodePattern.MatchString(code) {
			return nil, fmt.Errorf("payment provider code %q is invalid", code)
		}
		if name == "" {
			return nil, fmt.Errorf("payment provider %q name is required", code)
		}
		if _, exists := seen[code]; exists {
			return nil, fmt.Errorf("payment provider code %q is duplicated", code)
		}
		seen[code] = struct{}{}

		if item.Enabled && checkoutURL == "" {
			return nil, fmt.Errorf("payment provider %q checkout URL is required when enabled", code)
		}
		if checkoutURL != "" {
			if err := validateCommerceCheckoutURLTemplate(checkoutURL); err != nil {
				return nil, fmt.Errorf("payment provider %q checkout URL is invalid: %w", code, err)
			}
		}

		normalized = append(normalized, CommercePaymentProviderSetting{
			Code:        code,
			Name:        name,
			Description: description,
			Icon:        icon,
			CheckoutURL: checkoutURL,
			Enabled:     item.Enabled,
			SortOrder:   sortOrder,
		})
	}

	sort.SliceStable(normalized, func(i, j int) bool {
		if normalized[i].SortOrder != normalized[j].SortOrder {
			return normalized[i].SortOrder < normalized[j].SortOrder
		}
		return normalized[i].Code < normalized[j].Code
	})

	return normalized, nil
}

func ToCommercePaymentProviderOptions(items []CommercePaymentProviderSetting) []CommercePaymentProviderOption {
	if len(items) == 0 {
		return []CommercePaymentProviderOption{}
	}

	options := make([]CommercePaymentProviderOption, 0, len(items))
	for _, item := range items {
		if !item.Enabled {
			continue
		}
		options = append(options, CommercePaymentProviderOption{
			Code:        item.Code,
			Name:        item.Name,
			Description: item.Description,
			Icon:        item.Icon,
			SortOrder:   item.SortOrder,
		})
	}
	return options
}

func FindEnabledCommercePaymentProvider(items []CommercePaymentProviderSetting, code string) *CommercePaymentProviderSetting {
	code = strings.ToLower(strings.TrimSpace(code))
	if code == "" {
		return nil
	}
	for i := range items {
		if !items[i].Enabled || items[i].Code != code {
			continue
		}
		item := items[i]
		return &item
	}
	return nil
}

func BuildCommerceCheckoutURL(template string, input CommerceCheckoutURLTemplateInput) (string, error) {
	template = strings.TrimSpace(template)
	if template == "" {
		return "", fmt.Errorf("checkout URL is required")
	}

	replacer := strings.NewReplacer(
		"{order_id}", strconv.FormatInt(input.OrderID, 10),
		"{order_no}", strings.TrimSpace(input.OrderNo),
		"{user_id}", strconv.FormatInt(input.UserID, 10),
		"{product_id}", strconv.FormatInt(input.ProductID, 10),
		"{price_id}", strconv.FormatInt(input.PriceID, 10),
		"{amount}", strconv.FormatFloat(input.Amount, 'f', 2, 64),
		"{currency}", strings.ToUpper(strings.TrimSpace(input.Currency)),
		"{payment_provider}", strings.ToLower(strings.TrimSpace(input.PaymentProvider)),
	)

	result := strings.TrimSpace(replacer.Replace(template))
	if err := config.ValidateAbsoluteHTTPURL(result); err != nil {
		return "", err
	}
	return result, nil
}

func validateCommerceCheckoutURLTemplate(template string) error {
	_, err := BuildCommerceCheckoutURL(template, CommerceCheckoutURLTemplateInput{
		OrderID:         1,
		OrderNo:         "CO20250101000000TEST",
		UserID:          1,
		ProductID:       1,
		PriceID:         1,
		Amount:          9.99,
		Currency:        "CNY",
		PaymentProvider: "demo",
	})
	return err
}
