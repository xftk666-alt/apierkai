package service

import (
	"context"
	"crypto/subtle"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var (
	ErrCommerceCallbackSecretNotConfigured = infraerrors.Forbidden("COMMERCE_CALLBACK_SECRET_NOT_CONFIGURED", "commerce callback secret is not configured")
	ErrCommerceCallbackUnauthorized       = infraerrors.Unauthorized("COMMERCE_CALLBACK_UNAUTHORIZED", "invalid commerce callback secret")
)

type CommerceFeatureFlags struct {
	PurchaseEnabled     bool
	PurchaseMode        string
	MarketplaceEnabled  bool
	WalletEnabled       bool
	OrdersEnabled       bool
	LegacyPurchaseURL   string
	PaymentProviders    []CommercePaymentProviderOption
}

type CommerceFeatureService struct {
	settingService *SettingService
}

func NewCommerceFeatureService(settingService *SettingService) *CommerceFeatureService {
	return &CommerceFeatureService{settingService: settingService}
}

func (s *CommerceFeatureService) GetFeatureFlags(ctx context.Context) (*CommerceFeatureFlags, error) {
	settings, err := s.settingService.GetPublicSettings(ctx)
	if err != nil {
		return nil, err
	}

	paymentProviders, err := s.GetEnabledPaymentProviderOptions(ctx)
	if err != nil {
		return nil, err
	}

	return &CommerceFeatureFlags{
		PurchaseEnabled:    settings.PurchaseSubscriptionEnabled,
		PurchaseMode:       NormalizeCommercePurchaseMode(settings.NativePurchaseMode),
		MarketplaceEnabled: settings.NativeMarketplaceEnabled,
		WalletEnabled:      settings.NativeWalletEnabled,
		OrdersEnabled:      settings.NativeOrdersEnabled,
		LegacyPurchaseURL:  settings.PurchaseSubscriptionURL,
		PaymentProviders:   paymentProviders,
	}, nil
}

func (s *CommerceFeatureService) GetEnabledPaymentProviderSettings(ctx context.Context) ([]CommercePaymentProviderSetting, error) {
	if s == nil || s.settingService == nil {
		return []CommercePaymentProviderSetting{}, nil
	}

	items, err := s.settingService.GetCommercePaymentProviders(ctx)
	if err != nil {
		return nil, err
	}

	enabled := make([]CommercePaymentProviderSetting, 0, len(items))
	for _, item := range items {
		if item.Enabled {
			enabled = append(enabled, item)
		}
	}
	return enabled, nil
}

func (s *CommerceFeatureService) GetEnabledPaymentProviderOptions(ctx context.Context) ([]CommercePaymentProviderOption, error) {
	items, err := s.GetEnabledPaymentProviderSettings(ctx)
	if err != nil {
		return nil, err
	}
	return ToCommercePaymentProviderOptions(items), nil
}

func (s *CommerceFeatureService) ValidateCallbackSecret(ctx context.Context, provided string) error {
	if s == nil || s.settingService == nil {
		return ErrCommerceCallbackSecretNotConfigured
	}

	expected, err := s.settingService.GetCommerceCallbackSecret(ctx)
	if err != nil {
		return err
	}
	if expected == "" {
		return ErrCommerceCallbackSecretNotConfigured
	}

	provided = strings.TrimSpace(provided)
	if provided == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
		return ErrCommerceCallbackUnauthorized
	}
	return nil
}
