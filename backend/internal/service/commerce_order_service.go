package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

func NewCommerceOrderService(
	repo CommerceOrderRepository,
	catalogService *CommerceCatalogService,
	userRepo UserRepository,
	subscriptionService *SubscriptionService,
	billingCacheService *BillingCacheService,
	authCacheInvalidator APIKeyAuthCacheInvalidator,
	entClient *dbent.Client,
) *CommerceOrderService {
	return &CommerceOrderService{
		repo:                 repo,
		catalogService:       catalogService,
		userRepo:             userRepo,
		subscriptionService:  subscriptionService,
		billingCacheService:  billingCacheService,
		authCacheInvalidator: authCacheInvalidator,
		entClient:            entClient,
	}
}

func (s *CommerceOrderService) CreateOrder(ctx context.Context, userID int64, input *CommerceOrderCreateInput) (*CommerceOrderView, error) {
	if input == nil {
		return nil, ErrCommerceOrderSnapshotInvalid
	}

	flags, err := s.catalogService.requireMarketplaceEnabled(ctx)
	if err != nil {
		return nil, err
	}
	if flags != nil && !flags.OrdersEnabled {
		return nil, ErrCommerceOrdersDisabled
	}

	product, selectedPrice, snapshot, err := s.resolveOrderSnapshot(ctx, input.ProductID, input.PriceID)
	if err != nil {
		return nil, err
	}
	providerSetting, paymentProvider, err := s.resolveOrderPaymentProvider(ctx, input.PaymentProvider)
	if err != nil {
		return nil, err
	}
	orderNo, err := generateCommerceOrderNo()
	if err != nil {
		return nil, fmt.Errorf("generate order number: %w", err)
	}
	snapshotMap, err := commerceOrderSnapshotToMap(snapshot)
	if err != nil {
		return nil, fmt.Errorf("encode order snapshot: %w", err)
	}

	order, err := s.repo.CreateOrder(ctx, &CommerceOrder{
		OrderNo:         orderNo,
		UserID:          userID,
		ProductID:       product.ID,
		PriceID:         selectedPrice.ID,
		Status:          CommerceOrderStatusPending,
		PaymentStatus:   CommercePaymentStatusPending,
		PaymentProvider: paymentProvider,
		Amount:          selectedPrice.Amount,
		Currency:        normalizeCommerceCurrency(selectedPrice.Currency),
		Snapshot:        snapshotMap,
	})
	if err != nil {
		return nil, err
	}

	view, err := s.buildOrderView(order)
	if err != nil {
		return nil, err
	}
	if providerSetting != nil {
		view.PaymentAction, err = s.buildPaymentAction(providerSetting, order)
		if err != nil {
			return nil, err
		}
	}
	return view, nil
}

func (s *CommerceOrderService) ListUserOrders(ctx context.Context, userID int64, params pagination.PaginationParams) ([]CommerceOrderView, *pagination.PaginationResult, error) {
	if err := s.requireOrdersEnabled(ctx); err != nil {
		return nil, nil, err
	}

	orders, pager, err := s.repo.ListOrders(ctx, normalizeCommercePaginationParams(params), CommerceOrderListFilters{
		UserID: &userID,
	})
	if err != nil {
		return nil, nil, err
	}

	views, err := s.buildOrderViews(orders)
	return views, pager, err
}

func (s *CommerceOrderService) ListAdminOrders(ctx context.Context, params pagination.PaginationParams, filters CommerceOrderListFilters) ([]CommerceOrderView, *pagination.PaginationResult, error) {
	orders, pager, err := s.repo.ListOrders(ctx, normalizeCommercePaginationParams(params), filters)
	if err != nil {
		return nil, nil, err
	}

	views, err := s.buildAdminOrderViews(ctx, orders)
	return views, pager, err
}

func (s *CommerceOrderService) ListUserWalletLedgers(ctx context.Context, userID int64, params pagination.PaginationParams) ([]CommerceWalletLedgerView, *pagination.PaginationResult, error) {
	if err := s.requireWalletEnabled(ctx); err != nil {
		return nil, nil, err
	}

	items, pager, err := s.repo.ListWalletLedgers(ctx, normalizeCommercePaginationParams(params), CommerceWalletLedgerListFilters{
		UserID: &userID,
	})
	if err != nil {
		return nil, nil, err
	}
	return buildCommerceWalletLedgerViews(items), pager, nil
}

func (s *CommerceOrderService) ListAdminWalletLedgers(ctx context.Context, params pagination.PaginationParams, filters CommerceWalletLedgerListFilters) ([]CommerceWalletLedgerView, *pagination.PaginationResult, error) {
	items, pager, err := s.repo.ListWalletLedgers(ctx, normalizeCommercePaginationParams(params), filters)
	if err != nil {
		return nil, nil, err
	}
	return buildCommerceWalletLedgerViews(items), pager, nil
}

func (s *CommerceOrderService) HandlePaymentCallback(ctx context.Context, provider string, input *CommerceOrderPaymentCallbackInput) (*CommerceOrderView, error) {
	if input == nil {
		return nil, ErrCommerceOrderSnapshotInvalid
	}

	orderNo := strings.TrimSpace(input.OrderNo)
	if orderNo == "" {
		return nil, ErrCommerceOrderNoRequired
	}

	order, err := s.repo.GetOrderByOrderNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	if order.Status == CommerceOrderStatusCompleted || order.PaymentStatus == CommercePaymentStatusSucceeded || order.CompletedAt != nil {
		return s.buildOrderView(order)
	}

	view, err := s.ManualCompleteOrder(ctx, order.ID, 0, &CommerceOrderManualCompleteInput{
		PaymentProvider: provider,
		ProviderTradeNo: strings.TrimSpace(input.ProviderTradeNo),
		RequestPayload:  strings.TrimSpace(input.RequestPayload),
		CallbackPayload: strings.TrimSpace(input.CallbackPayload),
		PaidAmount:      input.PaidAmount,
		PaidCurrency:    strings.TrimSpace(input.PaidCurrency),
	})
	if err == nil {
		return view, nil
	}
	if !errors.Is(err, ErrCommerceOrderAlreadyHandled) {
		return nil, err
	}

	order, fetchErr := s.repo.GetOrderByID(ctx, order.ID)
	if fetchErr != nil {
		return nil, fetchErr
	}
	return s.buildAdminOrderView(ctx, order)
}

func (s *CommerceOrderService) resolveOrderPaymentProvider(ctx context.Context, provider string) (*CommercePaymentProviderSetting, string, error) {
	provider = strings.TrimSpace(provider)
	enabledProviders, err := s.catalogService.GetEnabledPaymentProviderSettings(ctx)
	if err != nil {
		return nil, "", err
	}

	if provider == "" {
		if len(enabledProviders) == 1 {
			return &enabledProviders[0], enabledProviders[0].Code, nil
		}
		return nil, CommercePaymentProviderManual, nil
	}

	normalized := normalizeCommercePaymentProvider(provider)
	if normalized == CommercePaymentProviderManual {
		return nil, normalized, nil
	}

	providerSetting := FindEnabledCommercePaymentProvider(enabledProviders, normalized)
	if providerSetting == nil {
		return nil, "", ErrCommercePaymentProviderNotFound
	}
	return providerSetting, providerSetting.Code, nil
}

func (s *CommerceOrderService) buildPaymentAction(provider *CommercePaymentProviderSetting, order *CommerceOrder) (*CommerceOrderPaymentAction, error) {
	if provider == nil || order == nil || strings.TrimSpace(provider.CheckoutURL) == "" {
		return nil, nil
	}

	checkoutURL, err := BuildCommerceCheckoutURL(provider.CheckoutURL, CommerceCheckoutURLTemplateInput{
		OrderID:         order.ID,
		OrderNo:         order.OrderNo,
		UserID:          order.UserID,
		ProductID:       order.ProductID,
		PriceID:         order.PriceID,
		Amount:          order.Amount,
		Currency:        order.Currency,
		PaymentProvider: provider.Code,
	})
	if err != nil {
		return nil, fmt.Errorf("build checkout url: %w", err)
	}

	return &CommerceOrderPaymentAction{
		Provider:     provider.Code,
		ProviderName: provider.Name,
		CheckoutURL:  checkoutURL,
	}, nil
}

func (s *CommerceOrderService) ManualCompleteOrder(ctx context.Context, orderID, adminID int64, input *CommerceOrderManualCompleteInput) (*CommerceOrderView, error) {
	tx, err := s.entClient.Tx(ctx)
	if err != nil && !errors.Is(err, dbent.ErrTxStarted) {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}

	txCtx := ctx
	if err == nil {
		txCtx = dbent.NewTxContext(ctx, tx)
		defer func() { _ = tx.Rollback() }()
	}

	order, err := s.repo.GetOrderByIDForUpdate(txCtx, orderID)
	if err != nil {
		return nil, err
	}
	if order.Status == CommerceOrderStatusCompleted || order.PaymentStatus == CommercePaymentStatusSucceeded || order.CompletedAt != nil {
		return nil, ErrCommerceOrderAlreadyHandled
	}

	snapshot, err := parseCommerceOrderSnapshot(order.Snapshot)
	if err != nil {
		return nil, err
	}
	if snapshot.Product.ProductType != CommerceProductTypeTopUpBalance && len(snapshot.Grants) == 0 {
		return nil, ErrCommerceOrderSnapshotInvalid
	}

	paymentProvider := order.PaymentProvider
	if input != nil && strings.TrimSpace(input.PaymentProvider) != "" {
		paymentProvider = input.PaymentProvider
	}
	paymentProvider = normalizeCommercePaymentProvider(paymentProvider)

	paidAmount := order.Amount
	if input != nil && input.PaidAmount != nil && *input.PaidAmount >= 0 {
		paidAmount = *input.PaidAmount
	}

	paidCurrency := order.Currency
	if input != nil && strings.TrimSpace(input.PaidCurrency) != "" {
		paidCurrency = normalizeCommerceCurrency(input.PaidCurrency)
	}

	paidAt := time.Now()
	invalidatedSubscriptionGroups := make([]int64, 0)
	shouldInvalidateBalance := false
	shouldInvalidateAuth := false

	if snapshot.Product.ProductType == CommerceProductTypeTopUpBalance {
		userBefore, err := s.userRepo.GetByID(txCtx, order.UserID)
		if err != nil {
			return nil, fmt.Errorf("get user before top-up: %w", err)
		}
		if err := s.userRepo.UpdateBalance(txCtx, order.UserID, paidAmount); err != nil {
			return nil, fmt.Errorf("update user balance: %w", err)
		}
		userAfter, err := s.userRepo.GetByID(txCtx, order.UserID)
		if err != nil {
			return nil, fmt.Errorf("get user after top-up: %w", err)
		}
		orderRef := order.ID
		if err := s.repo.CreateWalletLedger(txCtx, &CommerceWalletLedger{
			UserID:        order.UserID,
			OrderID:       &orderRef,
			Direction:     CommerceWalletDirectionCredit,
			ChangeAmount:  paidAmount,
			BalanceBefore: userBefore.Balance,
			BalanceAfter:  userAfter.Balance,
			ReasonType:    CommerceWalletReasonTypeOrderTopUp,
			ReasonDetail:  commerceOrderNotePrefix + order.OrderNo,
		}); err != nil {
			return nil, fmt.Errorf("create wallet ledger: %w", err)
		}
		shouldInvalidateBalance = true
		shouldInvalidateAuth = true
	}

	for _, grant := range snapshot.Grants {
		switch grant.GrantType {
		case CommerceGrantTypeSubscription:
			validityDays := grant.ValidityDays
			if validityDays <= 0 {
				validityDays = s.defaultGrantValidityDays(txCtx, grant.GroupID)
			}
			_, _, err := s.subscriptionService.AssignOrExtendSubscription(txCtx, &AssignSubscriptionInput{
				UserID:       order.UserID,
				GroupID:      grant.GroupID,
				ValidityDays: validityDays,
				AssignedBy:   adminID,
				Notes:        commerceOrderNotePrefix + order.OrderNo,
			})
			if err != nil {
				return nil, fmt.Errorf("grant subscription: %w", err)
			}
			invalidatedSubscriptionGroups = appendUniqueInt64(invalidatedSubscriptionGroups, grant.GroupID)
			shouldInvalidateAuth = true
		case CommerceGrantTypeAllowedGroup:
			if err := s.userRepo.AddGroupToAllowedGroups(txCtx, order.UserID, grant.GroupID); err != nil {
				return nil, fmt.Errorf("grant allowed group: %w", err)
			}
			shouldInvalidateAuth = true
		}
	}

	transaction := &CommercePaymentTransaction{
		OrderID:      order.ID,
		Provider:     paymentProvider,
		Status:       CommercePaymentStatusSucceeded,
		PaidAmount:   paidAmount,
		PaidCurrency: paidCurrency,
	}
	if input != nil {
		transaction.ProviderTradeNo = strings.TrimSpace(input.ProviderTradeNo)
		transaction.RequestPayload = strings.TrimSpace(input.RequestPayload)
		transaction.CallbackPayload = strings.TrimSpace(input.CallbackPayload)
	}
	if err := s.repo.CreatePaymentTransaction(txCtx, transaction); err != nil {
		return nil, fmt.Errorf("create payment transaction: %w", err)
	}

	if _, err := s.repo.CompleteOrder(txCtx, order.ID, &CommerceOrderCompleteRecord{
		Status:          CommerceOrderStatusCompleted,
		PaymentStatus:   CommercePaymentStatusSucceeded,
		PaymentProvider: paymentProvider,
		Amount:          paidAmount,
		Currency:        paidCurrency,
		PaidAt:          &paidAt,
		CompletedAt:     &paidAt,
	}); err != nil {
		return nil, fmt.Errorf("complete order: %w", err)
	}

	if tx != nil {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit transaction: %w", err)
		}
	}

	s.invalidateOrderCaches(ctx, order.UserID, shouldInvalidateAuth, shouldInvalidateBalance, invalidatedSubscriptionGroups)

	order, err = s.repo.GetOrderByID(ctx, order.ID)
	if err != nil {
		return nil, err
	}
	return s.buildOrderView(order)
}

func (s *CommerceOrderService) resolveOrderSnapshot(ctx context.Context, productID, priceID int64) (*CommerceProduct, *CommerceCatalogPrice, *CommerceOrderSnapshot, error) {
	products, prices, grants, err := s.catalogService.repo.GetProductsByIDs(ctx, []int64{productID})
	if err != nil {
		return nil, nil, nil, err
	}
	if len(products) == 0 {
		return nil, nil, nil, ErrCommerceProductNotFound
	}

	product := &products[0]
	if product.Status != CommerceProductStatusActive {
		return nil, nil, nil, ErrCommerceProductInactive
	}

	view, err := s.catalogService.buildCatalogProduct(ctx, product, prices, grants)
	if err != nil {
		return nil, nil, nil, err
	}

	var selectedPrice *CommerceCatalogPrice
	for i := range view.Prices {
		if view.Prices[i].ID == priceID {
			selectedPrice = &view.Prices[i]
			break
		}
	}
	if selectedPrice == nil {
		return nil, nil, nil, ErrCommercePriceNotFound
	}
	if !selectedPrice.Enabled {
		return nil, nil, nil, ErrCommercePriceDisabled
	}

	snapshot := &CommerceOrderSnapshot{
		Product: CommerceOrderSnapshotProduct{
			ID:          view.ID,
			Code:        view.Code,
			Name:        view.Name,
			Description: view.Description,
			ProductType: view.ProductType,
			Status:      view.Status,
			CoverImage:  view.CoverImage,
			Tags:        cloneStringSlice(view.Tags),
			SortOrder:   view.SortOrder,
			Recommended: view.Recommended,
			Metadata:    cloneMap(view.Metadata),
		},
		Price: CommerceCatalogPrice{
			ID:             selectedPrice.ID,
			PriceType:      selectedPrice.PriceType,
			Amount:         selectedPrice.Amount,
			Currency:       selectedPrice.Currency,
			OriginalAmount: selectedPrice.OriginalAmount,
			Enabled:        selectedPrice.Enabled,
			SortOrder:      selectedPrice.SortOrder,
		},
		Grants: cloneCommerceCatalogGrants(view.Grants),
	}
	return product, selectedPrice, snapshot, nil
}

func (s *CommerceOrderService) requireOrdersEnabled(ctx context.Context) error {
	flags, err := s.catalogService.getFeatureFlags(ctx)
	if err != nil {
		return err
	}
	if flags == nil || !flags.OrdersEnabled {
		return ErrCommerceOrdersDisabled
	}
	return nil
}

func (s *CommerceOrderService) requireWalletEnabled(ctx context.Context) error {
	flags, err := s.catalogService.getFeatureFlags(ctx)
	if err != nil {
		return err
	}
	if flags == nil || !flags.WalletEnabled {
		return ErrCommerceWalletDisabled
	}
	return nil
}

func (s *CommerceOrderService) defaultGrantValidityDays(ctx context.Context, groupID int64) int {
	if s.catalogService == nil || s.catalogService.groupRepo == nil || groupID <= 0 {
		return 30
	}
	group, err := s.catalogService.groupRepo.GetByID(ctx, groupID)
	if err != nil || group == nil || group.DefaultValidityDays <= 0 {
		return 30
	}
	return group.DefaultValidityDays
}

func (s *CommerceOrderService) invalidateOrderCaches(ctx context.Context, userID int64, invalidateAuth, invalidateBalance bool, subscriptionGroupIDs []int64) {
	if invalidateAuth && s.authCacheInvalidator != nil {
		s.authCacheInvalidator.InvalidateAuthCacheByUserID(ctx, userID)
	}
	if s.billingCacheService == nil {
		return
	}

	if invalidateBalance {
		go func() {
			cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = s.billingCacheService.InvalidateUserBalance(cacheCtx, userID)
		}()
	}

	for _, groupID := range subscriptionGroupIDs {
		groupID := groupID
		go func() {
			cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = s.billingCacheService.InvalidateSubscription(cacheCtx, userID, groupID)
		}()
	}
}

func (s *CommerceOrderService) buildOrderViews(orders []CommerceOrder) ([]CommerceOrderView, error) {
	views := make([]CommerceOrderView, 0, len(orders))
	for i := range orders {
		view, err := s.buildOrderView(&orders[i])
		if err != nil {
			return nil, err
		}
		views = append(views, *view)
	}
	return views, nil
}

func (s *CommerceOrderService) buildAdminOrderViews(ctx context.Context, orders []CommerceOrder) ([]CommerceOrderView, error) {
	views, err := s.buildOrderViews(orders)
	if err != nil || len(views) == 0 {
		return views, err
	}

	orderIDs := make([]int64, 0, len(orders))
	for _, order := range orders {
		orderIDs = append(orderIDs, order.ID)
	}

	transactionsByOrderID, err := s.repo.ListLatestPaymentTransactionsByOrderIDs(ctx, orderIDs)
	if err != nil {
		return nil, err
	}

	for i := range views {
		if transaction, ok := transactionsByOrderID[views[i].ID]; ok {
			views[i].PaymentTransaction = buildCommerceOrderPaymentTransactionView(&transaction)
		}
	}

	return views, nil
}

func (s *CommerceOrderService) buildAdminOrderView(ctx context.Context, order *CommerceOrder) (*CommerceOrderView, error) {
	if order == nil {
		return nil, ErrCommerceOrderNotFound
	}

	views, err := s.buildAdminOrderViews(ctx, []CommerceOrder{*order})
	if err != nil {
		return nil, err
	}
	if len(views) == 0 {
		return nil, ErrCommerceOrderNotFound
	}

	return &views[0], nil
}

func (s *CommerceOrderService) buildOrderView(order *CommerceOrder) (*CommerceOrderView, error) {
	if order == nil {
		return nil, ErrCommerceOrderNotFound
	}

	snapshot, err := parseCommerceOrderSnapshot(order.Snapshot)
	if err != nil {
		return nil, err
	}

	return &CommerceOrderView{
		ID:              order.ID,
		OrderNo:         order.OrderNo,
		UserID:          order.UserID,
		ProductID:       order.ProductID,
		PriceID:         order.PriceID,
		Status:          order.Status,
		PaymentStatus:   order.PaymentStatus,
		PaymentProvider: order.PaymentProvider,
		Amount:          order.Amount,
		Currency:        order.Currency,
		Snapshot:        *snapshot,
		PaidAt:          order.PaidAt,
		ExpiredAt:       order.ExpiredAt,
		CompletedAt:     order.CompletedAt,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
	}, nil
}

func buildCommerceOrderPaymentTransactionView(transaction *CommercePaymentTransaction) *CommerceOrderPaymentTransactionView {
	if transaction == nil {
		return nil
	}

	return &CommerceOrderPaymentTransactionView{
		ID:              transaction.ID,
		OrderID:         transaction.OrderID,
		Provider:        transaction.Provider,
		ProviderTradeNo: transaction.ProviderTradeNo,
		Status:          transaction.Status,
		RequestPayload:  transaction.RequestPayload,
		CallbackPayload: transaction.CallbackPayload,
		PaidAmount:      transaction.PaidAmount,
		PaidCurrency:    transaction.PaidCurrency,
		CreatedAt:       transaction.CreatedAt,
		UpdatedAt:       transaction.UpdatedAt,
	}
}

func buildCommerceWalletLedgerViews(items []CommerceWalletLedgerRecord) []CommerceWalletLedgerView {
	views := make([]CommerceWalletLedgerView, 0, len(items))
	for _, item := range items {
		views = append(views, CommerceWalletLedgerView{
			ID:            item.ID,
			UserID:        item.UserID,
			OrderID:       item.OrderID,
			OrderNo:       item.OrderNo,
			Direction:     item.Direction,
			ChangeAmount:  item.ChangeAmount,
			BalanceBefore: item.BalanceBefore,
			BalanceAfter:  item.BalanceAfter,
			ReasonType:    item.ReasonType,
			ReasonDetail:  item.ReasonDetail,
			CreatedAt:     item.CreatedAt,
			UpdatedAt:     item.UpdatedAt,
		})
	}
	return views
}

func parseCommerceOrderSnapshot(raw map[string]any) (*CommerceOrderSnapshot, error) {
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, ErrCommerceOrderSnapshotInvalid.WithCause(err)
	}

	var snapshot CommerceOrderSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, ErrCommerceOrderSnapshotInvalid.WithCause(err)
	}
	if snapshot.Product.ID <= 0 || snapshot.Price.ID <= 0 {
		return nil, ErrCommerceOrderSnapshotInvalid
	}
	return &snapshot, nil
}

func commerceOrderSnapshotToMap(snapshot *CommerceOrderSnapshot) (map[string]any, error) {
	data, err := json.Marshal(snapshot)
	if err != nil {
		return nil, err
	}

	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = map[string]any{}
	}
	return out, nil
}

func cloneCommerceCatalogGrants(items []CommerceCatalogGrant) []CommerceCatalogGrant {
	if len(items) == 0 {
		return []CommerceCatalogGrant{}
	}
	out := make([]CommerceCatalogGrant, len(items))
	copy(out, items)
	return out
}

func appendUniqueInt64(items []int64, value int64) []int64 {
	for _, item := range items {
		if item == value {
			return items
		}
	}
	return append(items, value)
}

func normalizeCommercePaymentProvider(provider string) string {
	provider = strings.TrimSpace(provider)
	if provider == "" {
		return CommercePaymentProviderManual
	}
	return provider
}

func normalizeCommercePaginationParams(params pagination.PaginationParams) pagination.PaginationParams {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 20
	}
	return params
}

func generateCommerceOrderNo() (string, error) {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "CO" + time.Now().UTC().Format("20060102150405") + strings.ToUpper(hex.EncodeToString(buf)), nil
}
