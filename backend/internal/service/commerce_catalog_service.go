package service

import (
	"context"
	"fmt"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var (
	ErrCommerceMarketplaceDisabled = infraerrors.Forbidden("COMMERCE_MARKETPLACE_DISABLED", "commerce marketplace is not enabled")
	ErrCommerceProductNotFound     = infraerrors.NotFound("COMMERCE_PRODUCT_NOT_FOUND", "commerce product not found")
	ErrCommerceProductCodeExists   = infraerrors.Conflict("COMMERCE_PRODUCT_CODE_EXISTS", "commerce product code already exists")
	ErrCommerceModelNotFound       = infraerrors.NotFound("COMMERCE_MODEL_NOT_FOUND", "commerce model listing not found")
	ErrCommerceModelKeyExists      = infraerrors.Conflict("COMMERCE_MODEL_KEY_EXISTS", "commerce model key already exists")
)

type CommerceCatalogRepository interface {
	ListProducts(ctx context.Context, includeInactive bool) ([]CommerceProduct, []CommerceProductPrice, []CommerceProductGroupBinding, error)
	GetProductsByIDs(ctx context.Context, ids []int64) ([]CommerceProduct, []CommerceProductPrice, []CommerceProductGroupBinding, error)
	ListModelListings(ctx context.Context, includeInactive bool) ([]CommerceModelListing, []CommerceModelProductBinding, error)
	ProductsExistByIDs(ctx context.Context, ids []int64) (map[int64]bool, error)
	UpsertProduct(ctx context.Context, id *int64, input *CommerceProductUpsertInput) (*CommerceProduct, []CommerceProductPrice, []CommerceProductGroupBinding, error)
	UpsertModelListing(ctx context.Context, id *int64, input *CommerceModelListingUpsertInput) (*CommerceModelListing, []CommerceModelProductBinding, error)
}

type CommerceCatalogService struct {
	repo           CommerceCatalogRepository
	groupRepo      GroupRepository
	featureService *CommerceFeatureService
}

func NewCommerceCatalogService(
	repo CommerceCatalogRepository,
	groupRepo GroupRepository,
	featureService *CommerceFeatureService,
) *CommerceCatalogService {
	return &CommerceCatalogService{
		repo:           repo,
		groupRepo:      groupRepo,
		featureService: featureService,
	}
}

func (s *CommerceCatalogService) GetUserCatalog(ctx context.Context) (*CommerceCatalog, error) {
	flags, err := s.requireMarketplaceEnabled(ctx)
	if err != nil {
		return nil, err
	}
	return s.buildCatalog(ctx, flags, false)
}

func (s *CommerceCatalogService) GetAdminCatalog(ctx context.Context) (*CommerceCatalog, error) {
	flags, err := s.getFeatureFlags(ctx)
	if err != nil {
		return nil, err
	}
	return s.buildCatalog(ctx, flags, true)
}

func (s *CommerceCatalogService) CreateProduct(ctx context.Context, input *CommerceProductUpsertInput) (*CommerceCatalogProduct, error) {
	normalized, err := s.normalizeProductInput(ctx, input)
	if err != nil {
		return nil, err
	}
	product, prices, grants, err := s.repo.UpsertProduct(ctx, nil, normalized)
	if err != nil {
		return nil, err
	}
	return s.buildCatalogProduct(ctx, product, prices, grants)
}

func (s *CommerceCatalogService) UpdateProduct(ctx context.Context, id int64, input *CommerceProductUpsertInput) (*CommerceCatalogProduct, error) {
	normalized, err := s.normalizeProductInput(ctx, input)
	if err != nil {
		return nil, err
	}
	product, prices, grants, err := s.repo.UpsertProduct(ctx, &id, normalized)
	if err != nil {
		return nil, err
	}
	return s.buildCatalogProduct(ctx, product, prices, grants)
}

func (s *CommerceCatalogService) CreateModelListing(ctx context.Context, input *CommerceModelListingUpsertInput) (*CommerceCatalogModel, error) {
	normalized, err := s.normalizeModelInput(ctx, input)
	if err != nil {
		return nil, err
	}
	model, bindings, err := s.repo.UpsertModelListing(ctx, nil, normalized)
	if err != nil {
		return nil, err
	}
	return s.buildCatalogModel(ctx, model, bindings)
}

func (s *CommerceCatalogService) UpdateModelListing(ctx context.Context, id int64, input *CommerceModelListingUpsertInput) (*CommerceCatalogModel, error) {
	normalized, err := s.normalizeModelInput(ctx, input)
	if err != nil {
		return nil, err
	}
	model, bindings, err := s.repo.UpsertModelListing(ctx, &id, normalized)
	if err != nil {
		return nil, err
	}
	return s.buildCatalogModel(ctx, model, bindings)
}

func (s *CommerceCatalogService) requireMarketplaceEnabled(ctx context.Context) (*CommerceFeatureFlags, error) {
	flags, err := s.getFeatureFlags(ctx)
	if err != nil {
		return nil, err
	}
	if !flags.PurchaseEnabled || flags.PurchaseMode != CommercePurchaseModeNative || !flags.MarketplaceEnabled {
		return nil, ErrCommerceMarketplaceDisabled
	}
	return flags, nil
}

func (s *CommerceCatalogService) getFeatureFlags(ctx context.Context) (*CommerceFeatureFlags, error) {
	if s.featureService == nil {
		return &CommerceFeatureFlags{PurchaseMode: CommercePurchaseModeIframe}, nil
	}
	return s.featureService.GetFeatureFlags(ctx)
}

func (s *CommerceCatalogService) GetEnabledPaymentProviderSettings(ctx context.Context) ([]CommercePaymentProviderSetting, error) {
	if s == nil || s.featureService == nil {
		return []CommercePaymentProviderSetting{}, nil
	}
	return s.featureService.GetEnabledPaymentProviderSettings(ctx)
}

func (s *CommerceCatalogService) ValidateCallbackSecret(ctx context.Context, provided string) error {
	if s == nil || s.featureService == nil {
		return ErrCommerceCallbackSecretNotConfigured
	}
	return s.featureService.ValidateCallbackSecret(ctx, provided)
}

func (s *CommerceCatalogService) buildCatalog(ctx context.Context, flags *CommerceFeatureFlags, includeInactive bool) (*CommerceCatalog, error) {
	products, prices, grants, err := s.repo.ListProducts(ctx, includeInactive)
	if err != nil {
		return nil, err
	}
	models, bindings, err := s.repo.ListModelListings(ctx, includeInactive)
	if err != nil {
		return nil, err
	}

	productViews, productByID, err := s.buildCatalogProducts(ctx, products, prices, grants)
	if err != nil {
		return nil, err
	}
	modelViews := make([]CommerceCatalogModel, 0, len(models))
	modelByID := make(map[int64]*CommerceCatalogModel, len(models))
	for _, item := range models {
		view := CommerceCatalogModel{
			ID:          item.ID,
			ModelKey:    item.ModelKey,
			DisplayName: item.DisplayName,
			Description: item.Description,
			Vendor:      item.Vendor,
			Icon:        item.Icon,
			Tags:        cloneStringSlice(item.Tags),
			Status:      item.Status,
			SortOrder:   item.SortOrder,
			Recommended: item.Recommended,
			Metadata:    cloneMap(item.Metadata),
			Offers:      []CommerceCatalogOffer{},
		}
		modelViews = append(modelViews, view)
		modelByID[item.ID] = &modelViews[len(modelViews)-1]
	}
	for _, binding := range bindings {
		model := modelByID[binding.ModelListingID]
		product := productByID[binding.ProductID]
		if model == nil || product == nil {
			continue
		}
		model.Offers = append(model.Offers, CommerceCatalogOffer{
			BindingID:   binding.ID,
			BindingType: binding.BindingType,
			Product:     *product,
		})
	}

	catalogFlags := CommerceCatalogFeatures{}
	if flags != nil {
		catalogFlags = CommerceCatalogFeatures{
			PurchaseEnabled:    flags.PurchaseEnabled,
			PurchaseMode:       flags.PurchaseMode,
			MarketplaceEnabled: flags.MarketplaceEnabled,
			WalletEnabled:      flags.WalletEnabled,
			OrdersEnabled:      flags.OrdersEnabled,
			LegacyPurchaseURL:  flags.LegacyPurchaseURL,
			PaymentProviders:   flags.PaymentProviders,
		}
	}

	return &CommerceCatalog{
		Features: catalogFlags,
		Products: productViews,
		Models:   modelViews,
	}, nil
}

func (s *CommerceCatalogService) buildCatalogProduct(
	ctx context.Context,
	product *CommerceProduct,
	prices []CommerceProductPrice,
	grants []CommerceProductGroupBinding,
) (*CommerceCatalogProduct, error) {
	views, byID, err := s.buildCatalogProducts(ctx, []CommerceProduct{*product}, prices, grants)
	if err != nil {
		return nil, err
	}
	if len(views) == 0 {
		return nil, ErrCommerceProductNotFound
	}
	view := byID[product.ID]
	if view == nil {
		return &views[0], nil
	}
	return view, nil
}

func (s *CommerceCatalogService) buildCatalogModel(
	ctx context.Context,
	model *CommerceModelListing,
	bindings []CommerceModelProductBinding,
) (*CommerceCatalogModel, error) {
	view := &CommerceCatalogModel{
		ID:          model.ID,
		ModelKey:    model.ModelKey,
		DisplayName: model.DisplayName,
		Description: model.Description,
		Vendor:      model.Vendor,
		Icon:        model.Icon,
		Tags:        cloneStringSlice(model.Tags),
		Status:      model.Status,
		SortOrder:   model.SortOrder,
		Recommended: model.Recommended,
		Metadata:    cloneMap(model.Metadata),
		Offers:      []CommerceCatalogOffer{},
	}
	productIDs := make([]int64, 0, len(bindings))
	for _, binding := range bindings {
		if binding.ProductID > 0 {
			productIDs = append(productIDs, binding.ProductID)
		}
	}
	if len(productIDs) == 0 {
		return view, nil
	}
	products, prices, grants, err := s.repo.GetProductsByIDs(ctx, productIDs)
	if err != nil {
		return nil, err
	}
	_, productByID, err := s.buildCatalogProducts(ctx, products, prices, grants)
	if err != nil {
		return nil, err
	}
	for _, binding := range bindings {
		product := productByID[binding.ProductID]
		if product == nil {
			continue
		}
		view.Offers = append(view.Offers, CommerceCatalogOffer{
			BindingID:   binding.ID,
			BindingType: binding.BindingType,
			Product:     *product,
		})
	}
	return view, nil
}

func (s *CommerceCatalogService) buildCatalogProducts(
	ctx context.Context,
	products []CommerceProduct,
	prices []CommerceProductPrice,
	grants []CommerceProductGroupBinding,
) ([]CommerceCatalogProduct, map[int64]*CommerceCatalogProduct, error) {
	groupNames, err := s.loadGroupNames(ctx, grants)
	if err != nil {
		return nil, nil, err
	}

	productViews := make([]CommerceCatalogProduct, 0, len(products))
	productByID := make(map[int64]*CommerceCatalogProduct, len(products))
	for _, item := range products {
		view := CommerceCatalogProduct{
			ID:          item.ID,
			Code:        item.Code,
			Name:        item.Name,
			Description: item.Description,
			ProductType: item.ProductType,
			Status:      item.Status,
			CoverImage:  item.CoverImage,
			Tags:        cloneStringSlice(item.Tags),
			SortOrder:   item.SortOrder,
			Recommended: item.Recommended,
			Metadata:    cloneMap(item.Metadata),
			Prices:      []CommerceCatalogPrice{},
			Grants:      []CommerceCatalogGrant{},
		}
		productViews = append(productViews, view)
		productByID[item.ID] = &productViews[len(productViews)-1]
	}
	for _, price := range prices {
		product := productByID[price.ProductID]
		if product == nil {
			continue
		}
		product.Prices = append(product.Prices, CommerceCatalogPrice{
			ID:             price.ID,
			PriceType:      price.PriceType,
			Amount:         price.Amount,
			Currency:       price.Currency,
			OriginalAmount: price.OriginalAmount,
			Enabled:        price.Enabled,
			SortOrder:      price.SortOrder,
		})
	}
	for _, grant := range grants {
		product := productByID[grant.ProductID]
		if product == nil {
			continue
		}
		product.Grants = append(product.Grants, CommerceCatalogGrant{
			ID:           grant.ID,
			GroupID:      grant.GroupID,
			GroupName:    groupNames[grant.GroupID],
			GrantType:    grant.GrantType,
			ValidityDays: grant.ValidityDays,
			Priority:     grant.Priority,
		})
	}
	return productViews, productByID, nil
}

func (s *CommerceCatalogService) loadGroupNames(ctx context.Context, grants []CommerceProductGroupBinding) (map[int64]string, error) {
	groupNames := make(map[int64]string)
	if s.groupRepo == nil {
		return groupNames, nil
	}
	seen := make(map[int64]struct{})
	for _, grant := range grants {
		if grant.GroupID <= 0 {
			continue
		}
		if _, ok := seen[grant.GroupID]; ok {
			continue
		}
		seen[grant.GroupID] = struct{}{}
		groupInfo, err := s.groupRepo.GetByID(ctx, grant.GroupID)
		if err != nil {
			if infraerrors.IsNotFound(err) {
				continue
			}
			return nil, err
		}
		groupNames[grant.GroupID] = groupInfo.Name
	}
	return groupNames, nil
}

func (s *CommerceCatalogService) normalizeProductInput(ctx context.Context, input *CommerceProductUpsertInput) (*CommerceProductUpsertInput, error) {
	if input == nil {
		return nil, infraerrors.BadRequest("COMMERCE_PRODUCT_INPUT_REQUIRED", "commerce product input is required")
	}

	normalized := &CommerceProductUpsertInput{
		Code:          strings.TrimSpace(input.Code),
		Name:          strings.TrimSpace(input.Name),
		Description:   strings.TrimSpace(input.Description),
		ProductType:   strings.TrimSpace(input.ProductType),
		Status:        normalizeCommerceStatus(input.Status),
		CoverImage:    strings.TrimSpace(input.CoverImage),
		Tags:          normalizeCommerceTags(input.Tags),
		SortOrder:     normalizeCommerceSortOrder(input.SortOrder),
		Recommended:   input.Recommended,
		Metadata:      cloneMap(input.Metadata),
		Prices:        make([]CommerceProductPriceInput, 0, len(input.Prices)),
		GroupBindings: make([]CommerceProductGrantInput, 0, len(input.GroupBindings)),
	}

	switch {
	case normalized.Code == "":
		return nil, infraerrors.BadRequest("COMMERCE_PRODUCT_CODE_REQUIRED", "product code is required")
	case normalized.Name == "":
		return nil, infraerrors.BadRequest("COMMERCE_PRODUCT_NAME_REQUIRED", "product name is required")
	}
	if !isAllowedCommerceProductType(normalized.ProductType) {
		return nil, infraerrors.BadRequest("COMMERCE_PRODUCT_TYPE_INVALID", "product_type is invalid")
	}

	groupIDs := make([]int64, 0, len(input.GroupBindings))
	for _, item := range input.Prices {
		price, err := normalizeCommercePriceInput(item)
		if err != nil {
			return nil, err
		}
		normalized.Prices = append(normalized.Prices, price)
	}
	for _, item := range input.GroupBindings {
		grant, err := normalizeCommerceGrantInput(item)
		if err != nil {
			return nil, err
		}
		groupIDs = append(groupIDs, grant.GroupID)
		normalized.GroupBindings = append(normalized.GroupBindings, grant)
	}
	if err := s.ensureGroupsExist(ctx, groupIDs); err != nil {
		return nil, err
	}

	return normalized, nil
}

func (s *CommerceCatalogService) normalizeModelInput(ctx context.Context, input *CommerceModelListingUpsertInput) (*CommerceModelListingUpsertInput, error) {
	if input == nil {
		return nil, infraerrors.BadRequest("COMMERCE_MODEL_INPUT_REQUIRED", "commerce model input is required")
	}

	normalized := &CommerceModelListingUpsertInput{
		ModelKey:        strings.TrimSpace(input.ModelKey),
		DisplayName:     strings.TrimSpace(input.DisplayName),
		Description:     strings.TrimSpace(input.Description),
		Vendor:          strings.TrimSpace(input.Vendor),
		Icon:            strings.TrimSpace(input.Icon),
		Tags:            normalizeCommerceTags(input.Tags),
		Status:          normalizeCommerceStatus(input.Status),
		SortOrder:       normalizeCommerceSortOrder(input.SortOrder),
		Recommended:     input.Recommended,
		Metadata:        cloneMap(input.Metadata),
		ProductBindings: make([]CommerceModelProductBindingInput, 0, len(input.ProductBindings)),
	}

	switch {
	case normalized.ModelKey == "":
		return nil, infraerrors.BadRequest("COMMERCE_MODEL_KEY_REQUIRED", "model_key is required")
	case normalized.DisplayName == "":
		return nil, infraerrors.BadRequest("COMMERCE_MODEL_DISPLAY_NAME_REQUIRED", "display_name is required")
	}

	productIDs := make([]int64, 0, len(input.ProductBindings))
	for _, item := range input.ProductBindings {
		binding, err := normalizeCommerceModelBindingInput(item)
		if err != nil {
			return nil, err
		}
		productIDs = append(productIDs, binding.ProductID)
		normalized.ProductBindings = append(normalized.ProductBindings, binding)
	}
	if err := s.ensureProductsExist(ctx, productIDs); err != nil {
		return nil, err
	}

	return normalized, nil
}

func (s *CommerceCatalogService) ensureGroupsExist(ctx context.Context, groupIDs []int64) error {
	if s.groupRepo == nil {
		return nil
	}
	seen := make(map[int64]struct{})
	for _, groupID := range groupIDs {
		if groupID <= 0 {
			return infraerrors.BadRequest("COMMERCE_GROUP_ID_INVALID", "group_id must be greater than 0")
		}
		if _, ok := seen[groupID]; ok {
			continue
		}
		seen[groupID] = struct{}{}
		if _, err := s.groupRepo.GetByID(ctx, groupID); err != nil {
			if infraerrors.IsNotFound(err) {
				return infraerrors.BadRequest("COMMERCE_GROUP_NOT_FOUND", fmt.Sprintf("group_id %d does not exist", groupID))
			}
			return err
		}
	}
	return nil
}

func (s *CommerceCatalogService) ensureProductsExist(ctx context.Context, productIDs []int64) error {
	if len(productIDs) == 0 {
		return nil
	}
	seen := make(map[int64]struct{})
	uniqueIDs := make([]int64, 0, len(productIDs))
	for _, productID := range productIDs {
		if productID <= 0 {
			return infraerrors.BadRequest("COMMERCE_PRODUCT_BINDING_ID_INVALID", "product_id must be greater than 0")
		}
		if _, ok := seen[productID]; ok {
			continue
		}
		seen[productID] = struct{}{}
		uniqueIDs = append(uniqueIDs, productID)
	}

	exists, err := s.repo.ProductsExistByIDs(ctx, uniqueIDs)
	if err != nil {
		return err
	}
	for _, productID := range uniqueIDs {
		if exists[productID] {
			continue
		}
		return infraerrors.BadRequest("COMMERCE_PRODUCT_BINDING_NOT_FOUND", fmt.Sprintf("product_id %d does not exist", productID))
	}
	return nil
}

func normalizeCommercePriceInput(input CommerceProductPriceInput) (CommerceProductPriceInput, error) {
	normalized := CommerceProductPriceInput{
		PriceType:      strings.TrimSpace(input.PriceType),
		Amount:         input.Amount,
		Currency:       normalizeCommerceCurrency(input.Currency),
		OriginalAmount: input.OriginalAmount,
		Enabled:        input.Enabled,
		SortOrder:      normalizeCommerceSortOrder(input.SortOrder),
	}
	if !isAllowedCommercePriceType(normalized.PriceType) {
		return CommerceProductPriceInput{}, infraerrors.BadRequest("COMMERCE_PRICE_TYPE_INVALID", "price_type is invalid")
	}
	if normalized.Amount < 0 {
		return CommerceProductPriceInput{}, infraerrors.BadRequest("COMMERCE_PRICE_AMOUNT_INVALID", "amount must be greater than or equal to 0")
	}
	if normalized.OriginalAmount != nil && *normalized.OriginalAmount < 0 {
		return CommerceProductPriceInput{}, infraerrors.BadRequest("COMMERCE_PRICE_ORIGINAL_AMOUNT_INVALID", "original_amount must be greater than or equal to 0")
	}
	return normalized, nil
}

func normalizeCommerceGrantInput(input CommerceProductGrantInput) (CommerceProductGrantInput, error) {
	normalized := CommerceProductGrantInput{
		GroupID:      input.GroupID,
		GrantType:    strings.TrimSpace(input.GrantType),
		ValidityDays: input.ValidityDays,
		Priority:     input.Priority,
	}
	if !isAllowedCommerceGrantType(normalized.GrantType) {
		return CommerceProductGrantInput{}, infraerrors.BadRequest("COMMERCE_GRANT_TYPE_INVALID", "grant_type is invalid")
	}
	if normalized.ValidityDays < 0 {
		return CommerceProductGrantInput{}, infraerrors.BadRequest("COMMERCE_VALIDITY_DAYS_INVALID", "validity_days must be greater than or equal to 0")
	}
	if normalized.Priority == 0 {
		normalized.Priority = 50
	}
	return normalized, nil
}

func normalizeCommerceModelBindingInput(input CommerceModelProductBindingInput) (CommerceModelProductBindingInput, error) {
	normalized := CommerceModelProductBindingInput{
		ProductID:   input.ProductID,
		BindingType: strings.TrimSpace(input.BindingType),
	}
	if !isAllowedCommerceBindingType(normalized.BindingType) {
		return CommerceModelProductBindingInput{}, infraerrors.BadRequest("COMMERCE_BINDING_TYPE_INVALID", "binding_type is invalid")
	}
	return normalized, nil
}

func normalizeCommerceCurrency(currency string) string {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		return "CNY"
	}
	return currency
}

func normalizeCommerceStatus(status string) string {
	status = strings.TrimSpace(status)
	switch status {
	case CommerceProductStatusDraft, CommerceProductStatusActive, CommerceProductStatusDisabled:
		return status
	default:
		return CommerceProductStatusActive
	}
}

func normalizeCommerceSortOrder(sortOrder int) int {
	if sortOrder == 0 {
		return 100
	}
	return sortOrder
}

func normalizeCommerceTags(tags []string) []string {
	if len(tags) == 0 {
		return []string{}
	}
	result := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, item := range tags {
		tag := strings.TrimSpace(item)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		result = append(result, tag)
	}
	return result
}

func isAllowedCommerceProductType(productType string) bool {
	switch productType {
	case CommerceProductTypeTopUpBalance,
		CommerceProductTypeSubscription,
		CommerceProductTypeStandardAccess,
		CommerceProductTypeCombo:
		return true
	default:
		return false
	}
}

func isAllowedCommercePriceType(priceType string) bool {
	switch priceType {
	case CommercePriceTypeOneTime,
		CommercePriceTypeMonthly,
		CommercePriceTypeQuarterly,
		CommercePriceTypeYearly:
		return true
	default:
		return false
	}
}

func isAllowedCommerceGrantType(grantType string) bool {
	switch grantType {
	case CommerceGrantTypeSubscription, CommerceGrantTypeAllowedGroup:
		return true
	default:
		return false
	}
}

func isAllowedCommerceBindingType(bindingType string) bool {
	switch bindingType {
	case CommerceModelBindingTypePrimary, CommerceModelBindingTypeUpsell, CommerceModelBindingTypeTopUp:
		return true
	default:
		return false
	}
}

func cloneStringSlice(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	out := make([]string, len(items))
	copy(out, items)
	return out
}

func cloneMap(input map[string]any) map[string]any {
	if len(input) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}
