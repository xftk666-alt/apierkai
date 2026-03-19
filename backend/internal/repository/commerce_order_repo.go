package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

const commerceOrderSelectColumns = `
	o.id,
	o.order_no,
	o.user_id,
	o.product_id,
	o.price_id,
	o.status,
	o.payment_status,
	o.payment_provider,
	o.amount::double precision,
	o.currency,
	o.snapshot,
	o.paid_at,
	o.expired_at,
	o.completed_at,
	o.created_at,
	o.updated_at
`

type commerceOrderRepository struct {
	sql sqlExecutor
}

func NewCommerceOrderRepository(sqlDB *sql.DB) service.CommerceOrderRepository {
	return &commerceOrderRepository{sql: sqlDB}
}

func (r *commerceOrderRepository) CreateOrder(ctx context.Context, order *service.CommerceOrder) (*service.CommerceOrder, error) {
	if order == nil {
		return nil, nil
	}

	snapshotJSON, err := json.Marshal(order.Snapshot)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
		INSERT INTO commerce_orders (
			order_no, user_id, product_id, price_id, status, payment_status, payment_provider, amount, currency, snapshot, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING %s
	`, commerceOrderSelectColumns)

	return r.queryOrderOne(
		ctx,
		query,
		order.OrderNo,
		order.UserID,
		order.ProductID,
		order.PriceID,
		order.Status,
		order.PaymentStatus,
		order.PaymentProvider,
		order.Amount,
		order.Currency,
		snapshotJSON,
	)
}

func (r *commerceOrderRepository) GetOrderByID(ctx context.Context, id int64) (*service.CommerceOrder, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM commerce_orders o
		WHERE o.id = $1 AND o.deleted_at IS NULL
	`, commerceOrderSelectColumns)
	return r.queryOrderOne(ctx, query, id)
}

func (r *commerceOrderRepository) GetOrderByOrderNo(ctx context.Context, orderNo string) (*service.CommerceOrder, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM commerce_orders o
		WHERE o.order_no = $1 AND o.deleted_at IS NULL
	`, commerceOrderSelectColumns)
	return r.queryOrderOne(ctx, query, strings.TrimSpace(orderNo))
}

func (r *commerceOrderRepository) GetOrderByIDForUpdate(ctx context.Context, id int64) (*service.CommerceOrder, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM commerce_orders o
		WHERE o.id = $1 AND o.deleted_at IS NULL
		FOR UPDATE
	`, commerceOrderSelectColumns)
	return r.queryOrderOne(ctx, query, id)
}

func (r *commerceOrderRepository) ListOrders(ctx context.Context, params pagination.PaginationParams, filters service.CommerceOrderListFilters) ([]service.CommerceOrder, *pagination.PaginationResult, error) {
	sqlq := r.executor(ctx)
	whereSQL, args := buildCommerceOrderWhere(filters)

	var total int64
	if err := scanSingleRow(ctx, sqlq, `SELECT COUNT(*) FROM commerce_orders o `+whereSQL, args, &total); err != nil {
		return nil, nil, err
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM commerce_orders o
		%s
		ORDER BY o.created_at DESC, o.id DESC
		LIMIT $%d OFFSET $%d
	`, commerceOrderSelectColumns, whereSQL, len(args)+1, len(args)+2)
	queryArgs := append(cloneArgs(args), params.Limit(), params.Offset())

	rows, err := sqlq.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]service.CommerceOrder, 0)
	for rows.Next() {
		item, err := scanCommerceOrder(rows)
		if err != nil {
			return nil, nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return items, paginationResultFromTotal(total, params), nil
}

func (r *commerceOrderRepository) ListLatestPaymentTransactionsByOrderIDs(ctx context.Context, orderIDs []int64) (map[int64]service.CommercePaymentTransaction, error) {
	result := make(map[int64]service.CommercePaymentTransaction)
	if len(orderIDs) == 0 {
		return result, nil
	}

	rows, err := r.executor(ctx).QueryContext(ctx, `
		SELECT DISTINCT ON (t.order_id)
			t.id,
			t.order_id,
			t.provider,
			t.provider_trade_no,
			t.status,
			t.request_payload,
			t.callback_payload,
			t.paid_amount::double precision,
			t.paid_currency,
			t.created_at,
			t.updated_at
		FROM commerce_payment_transactions t
		WHERE t.order_id = ANY($1) AND t.deleted_at IS NULL
		ORDER BY t.order_id, t.created_at DESC, t.id DESC
	`, orderIDs)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		item, err := scanCommercePaymentTransaction(rows)
		if err != nil {
			return nil, err
		}
		result[item.OrderID] = item
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *commerceOrderRepository) CompleteOrder(ctx context.Context, id int64, input *service.CommerceOrderCompleteRecord) (*service.CommerceOrder, error) {
	if input == nil {
		return nil, nil
	}

	query := fmt.Sprintf(`
		UPDATE commerce_orders
		SET status = $2,
			payment_status = $3,
			payment_provider = $4,
			amount = $5,
			currency = $6,
			paid_at = $7,
			completed_at = $8,
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING %s
	`, commerceOrderSelectColumns)

	return r.queryOrderOne(
		ctx,
		query,
		id,
		input.Status,
		input.PaymentStatus,
		input.PaymentProvider,
		input.Amount,
		input.Currency,
		input.PaidAt,
		input.CompletedAt,
	)
}

func (r *commerceOrderRepository) CreatePaymentTransaction(ctx context.Context, transaction *service.CommercePaymentTransaction) error {
	if transaction == nil {
		return nil
	}

	_, err := r.executor(ctx).ExecContext(ctx, `
		INSERT INTO commerce_payment_transactions (
			order_id, provider, provider_trade_no, status, request_payload, callback_payload, paid_amount, paid_currency, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
	`,
		transaction.OrderID,
		transaction.Provider,
		transaction.ProviderTradeNo,
		transaction.Status,
		transaction.RequestPayload,
		transaction.CallbackPayload,
		transaction.PaidAmount,
		transaction.PaidCurrency,
	)
	return err
}

func (r *commerceOrderRepository) CreateWalletLedger(ctx context.Context, ledger *service.CommerceWalletLedger) error {
	if ledger == nil {
		return nil
	}

	_, err := r.executor(ctx).ExecContext(ctx, `
		INSERT INTO commerce_wallet_ledgers (
			user_id, order_id, direction, change_amount, balance_before, balance_after, reason_type, reason_detail, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
	`,
		ledger.UserID,
		ledger.OrderID,
		ledger.Direction,
		ledger.ChangeAmount,
		ledger.BalanceBefore,
		ledger.BalanceAfter,
		ledger.ReasonType,
		ledger.ReasonDetail,
	)
	return err
}

func (r *commerceOrderRepository) ListWalletLedgers(ctx context.Context, params pagination.PaginationParams, filters service.CommerceWalletLedgerListFilters) ([]service.CommerceWalletLedgerRecord, *pagination.PaginationResult, error) {
	sqlq := r.executor(ctx)
	whereSQL, args := buildCommerceWalletLedgerWhere(filters)

	var total int64
	countQuery := `
		SELECT COUNT(*)
		FROM commerce_wallet_ledgers l
		LEFT JOIN commerce_orders o ON o.id = l.order_id AND o.deleted_at IS NULL
	` + whereSQL
	if err := scanSingleRow(ctx, sqlq, countQuery, args, &total); err != nil {
		return nil, nil, err
	}

	query := fmt.Sprintf(`
		SELECT
			l.id,
			l.user_id,
			l.order_id,
			COALESCE(o.order_no, ''),
			l.direction,
			l.change_amount::double precision,
			l.balance_before::double precision,
			l.balance_after::double precision,
			l.reason_type,
			l.reason_detail,
			l.created_at,
			l.updated_at
		FROM commerce_wallet_ledgers l
		LEFT JOIN commerce_orders o ON o.id = l.order_id AND o.deleted_at IS NULL
		%s
		ORDER BY l.created_at DESC, l.id DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, len(args)+1, len(args)+2)
	queryArgs := append(cloneArgs(args), params.Limit(), params.Offset())

	rows, err := sqlq.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]service.CommerceWalletLedgerRecord, 0)
	for rows.Next() {
		item, err := scanCommerceWalletLedgerRecord(rows)
		if err != nil {
			return nil, nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return items, paginationResultFromTotal(total, params), nil
}

func (r *commerceOrderRepository) executor(ctx context.Context) sqlExecutor {
	if tx := dbent.TxFromContext(ctx); tx != nil {
		return tx.Client()
	}
	return r.sql
}

func (r *commerceOrderRepository) queryOrderOne(ctx context.Context, query string, args ...any) (*service.CommerceOrder, error) {
	rows, err := r.executor(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, translatePersistenceError(err, nil, nil)
	}
	defer func() { _ = rows.Close() }()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, service.ErrCommerceOrderNotFound
	}

	item, err := scanCommerceOrder(rows)
	if err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &item, nil
}

func scanCommerceOrder(scanner interface{ Scan(dest ...any) error }) (service.CommerceOrder, error) {
	var (
		item        service.CommerceOrder
		snapshotRaw []byte
		paidAt      sql.NullTime
		expiredAt   sql.NullTime
		completedAt sql.NullTime
	)

	if err := scanner.Scan(
		&item.ID,
		&item.OrderNo,
		&item.UserID,
		&item.ProductID,
		&item.PriceID,
		&item.Status,
		&item.PaymentStatus,
		&item.PaymentProvider,
		&item.Amount,
		&item.Currency,
		&snapshotRaw,
		&paidAt,
		&expiredAt,
		&completedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return service.CommerceOrder{}, err
	}

	item.Snapshot = decodeMapJSON(snapshotRaw)
	if paidAt.Valid {
		value := paidAt.Time
		item.PaidAt = &value
	}
	if expiredAt.Valid {
		value := expiredAt.Time
		item.ExpiredAt = &value
	}
	if completedAt.Valid {
		value := completedAt.Time
		item.CompletedAt = &value
	}
	return item, nil
}

func scanCommercePaymentTransaction(scanner interface{ Scan(dest ...any) error }) (service.CommercePaymentTransaction, error) {
	var item service.CommercePaymentTransaction

	if err := scanner.Scan(
		&item.ID,
		&item.OrderID,
		&item.Provider,
		&item.ProviderTradeNo,
		&item.Status,
		&item.RequestPayload,
		&item.CallbackPayload,
		&item.PaidAmount,
		&item.PaidCurrency,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return service.CommercePaymentTransaction{}, err
	}

	return item, nil
}

func scanCommerceWalletLedgerRecord(scanner interface{ Scan(dest ...any) error }) (service.CommerceWalletLedgerRecord, error) {
	var (
		record  service.CommerceWalletLedgerRecord
		orderID sql.NullInt64
	)

	if err := scanner.Scan(
		&record.ID,
		&record.UserID,
		&orderID,
		&record.OrderNo,
		&record.Direction,
		&record.ChangeAmount,
		&record.BalanceBefore,
		&record.BalanceAfter,
		&record.ReasonType,
		&record.ReasonDetail,
		&record.CreatedAt,
		&record.UpdatedAt,
	); err != nil {
		return service.CommerceWalletLedgerRecord{}, err
	}

	if orderID.Valid {
		value := orderID.Int64
		record.OrderID = &value
	}
	return record, nil
}

func buildCommerceOrderWhere(filters service.CommerceOrderListFilters) (string, []any) {
	conditions := []string{"o.deleted_at IS NULL"}
	args := make([]any, 0, 4)

	if filters.UserID != nil && *filters.UserID > 0 {
		args = append(args, *filters.UserID)
		conditions = append(conditions, fmt.Sprintf("o.user_id = $%d", len(args)))
	}
	if status := strings.TrimSpace(filters.Status); status != "" {
		args = append(args, status)
		conditions = append(conditions, fmt.Sprintf("o.status = $%d", len(args)))
	}
	if paymentStatus := strings.TrimSpace(filters.PaymentStatus); paymentStatus != "" {
		args = append(args, paymentStatus)
		conditions = append(conditions, fmt.Sprintf("o.payment_status = $%d", len(args)))
	}
	if orderNo := strings.TrimSpace(filters.OrderNo); orderNo != "" {
		args = append(args, "%"+orderNo+"%")
		conditions = append(conditions, fmt.Sprintf("o.order_no ILIKE $%d", len(args)))
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

func buildCommerceWalletLedgerWhere(filters service.CommerceWalletLedgerListFilters) (string, []any) {
	conditions := []string{"l.deleted_at IS NULL"}
	args := make([]any, 0, 4)

	if filters.UserID != nil && *filters.UserID > 0 {
		args = append(args, *filters.UserID)
		conditions = append(conditions, fmt.Sprintf("l.user_id = $%d", len(args)))
	}
	if filters.OrderID != nil && *filters.OrderID > 0 {
		args = append(args, *filters.OrderID)
		conditions = append(conditions, fmt.Sprintf("l.order_id = $%d", len(args)))
	}
	if direction := strings.TrimSpace(filters.Direction); direction != "" {
		args = append(args, direction)
		conditions = append(conditions, fmt.Sprintf("l.direction = $%d", len(args)))
	}
	if reasonType := strings.TrimSpace(filters.ReasonType); reasonType != "" {
		args = append(args, reasonType)
		conditions = append(conditions, fmt.Sprintf("l.reason_type = $%d", len(args)))
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}

func cloneArgs(args []any) []any {
	if len(args) == 0 {
		return nil
	}

	out := make([]any, len(args))
	copy(out, args)
	return out
}
