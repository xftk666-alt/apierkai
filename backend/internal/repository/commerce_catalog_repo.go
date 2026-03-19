package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

type commerceCatalogRepository struct {
	db *sql.DB
}

func NewCommerceCatalogRepository(db *sql.DB) service.CommerceCatalogRepository {
	return &commerceCatalogRepository{db: db}
}

func (r *commerceCatalogRepository) ListProducts(ctx context.Context, includeInactive bool) ([]service.CommerceProduct, []service.CommerceProductPrice, []service.CommerceProductGroupBinding, error) {
	return r.selectProducts(ctx, nil, includeInactive)
}

func (r *commerceCatalogRepository) GetProductsByIDs(ctx context.Context, ids []int64) ([]service.CommerceProduct, []service.CommerceProductPrice, []service.CommerceProductGroupBinding, error) {
	return r.selectProducts(ctx, ids, true)
}

func (r *commerceCatalogRepository) ListModelListings(ctx context.Context, includeInactive bool) ([]service.CommerceModelListing, []service.CommerceModelProductBinding, error) {
	return r.selectModelListings(ctx, nil, includeInactive)
}

func (r *commerceCatalogRepository) ProductsExistByIDs(ctx context.Context, ids []int64) (map[int64]bool, error) {
	result := make(map[int64]bool, len(ids))
	if len(ids) == 0 {
		return result, nil
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id
		FROM commerce_products
		WHERE deleted_at IS NULL AND id = ANY($1)
	`, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result[id] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *commerceCatalogRepository) UpsertProduct(
	ctx context.Context,
	id *int64,
	input *service.CommerceProductUpsertInput,
) (*service.CommerceProduct, []service.CommerceProductPrice, []service.CommerceProductGroupBinding, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, nil, err
	}
	defer func() { _ = tx.Rollback() }()

	productID, err := upsertCommerceProductRow(ctx, tx, id, input)
	if err != nil {
		return nil, nil, nil, err
	}
	if err := replaceCommerceProductPrices(ctx, tx, productID, input.Prices); err != nil {
		return nil, nil, nil, err
	}
	if err := replaceCommerceProductGroupBindings(ctx, tx, productID, input.GroupBindings); err != nil {
		return nil, nil, nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, nil, nil, err
	}

	products, prices, grants, err := r.GetProductsByIDs(ctx, []int64{productID})
	if err != nil {
		return nil, nil, nil, err
	}
	if len(products) == 0 {
		return nil, nil, nil, service.ErrCommerceProductNotFound
	}
	return &products[0], prices, grants, nil
}

func (r *commerceCatalogRepository) UpsertModelListing(
	ctx context.Context,
	id *int64,
	input *service.CommerceModelListingUpsertInput,
) (*service.CommerceModelListing, []service.CommerceModelProductBinding, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = tx.Rollback() }()

	modelID, err := upsertCommerceModelRow(ctx, tx, id, input)
	if err != nil {
		return nil, nil, err
	}
	if err := replaceCommerceModelBindings(ctx, tx, modelID, input.ProductBindings); err != nil {
		return nil, nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}

	models, bindings, err := r.selectModelListings(ctx, []int64{modelID}, true)
	if err != nil {
		return nil, nil, err
	}
	if len(models) == 0 {
		return nil, nil, service.ErrCommerceModelNotFound
	}
	return &models[0], bindings, nil
}

func (r *commerceCatalogRepository) selectProducts(
	ctx context.Context,
	ids []int64,
	includeInactive bool,
) ([]service.CommerceProduct, []service.CommerceProductPrice, []service.CommerceProductGroupBinding, error) {
	query := `
		SELECT id, code, name, description, product_type, status, cover_image, tags, sort_order, recommended, metadata, created_at, updated_at
		FROM commerce_products
		WHERE deleted_at IS NULL
	`
	args := make([]any, 0, 2)
	if !includeInactive {
		query += ` AND status = $1`
		args = append(args, service.CommerceProductStatusActive)
	}
	if len(ids) > 0 {
		args = append(args, pq.Array(ids))
		query += fmt.Sprintf(` AND id = ANY($%d)`, len(args))
	}
	query += ` ORDER BY sort_order ASC, id ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, nil, err
	}
	defer func() { _ = rows.Close() }()

	products := make([]service.CommerceProduct, 0)
	productIDs := make([]int64, 0)
	for rows.Next() {
		var (
			item        service.CommerceProduct
			tagsRaw     []byte
			metadataRaw []byte
		)
		if err := rows.Scan(
			&item.ID,
			&item.Code,
			&item.Name,
			&item.Description,
			&item.ProductType,
			&item.Status,
			&item.CoverImage,
			&tagsRaw,
			&item.SortOrder,
			&item.Recommended,
			&metadataRaw,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, nil, nil, err
		}
		item.Tags = decodeStringSliceJSON(tagsRaw)
		item.Metadata = decodeMapJSON(metadataRaw)
		products = append(products, item)
		productIDs = append(productIDs, item.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, nil, err
	}
	if len(productIDs) == 0 {
		return products, []service.CommerceProductPrice{}, []service.CommerceProductGroupBinding{}, nil
	}

	prices, err := r.selectProductPrices(ctx, productIDs, includeInactive)
	if err != nil {
		return nil, nil, nil, err
	}
	grants, err := r.selectProductGrants(ctx, productIDs)
	if err != nil {
		return nil, nil, nil, err
	}

	return products, prices, grants, nil
}

func (r *commerceCatalogRepository) selectProductPrices(ctx context.Context, productIDs []int64, includeInactive bool) ([]service.CommerceProductPrice, error) {
	query := `
		SELECT id, product_id, price_type, amount::double precision, currency, original_amount::double precision, enabled, sort_order, created_at, updated_at
		FROM commerce_product_prices
		WHERE deleted_at IS NULL AND product_id = ANY($1)
	`
	args := []any{pq.Array(productIDs)}
	if !includeInactive {
		query += ` AND enabled = TRUE`
	}
	query += ` ORDER BY product_id ASC, sort_order ASC, id ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	prices := make([]service.CommerceProductPrice, 0)
	for rows.Next() {
		var (
			item           service.CommerceProductPrice
			originalAmount sql.NullFloat64
		)
		if err := rows.Scan(
			&item.ID,
			&item.ProductID,
			&item.PriceType,
			&item.Amount,
			&item.Currency,
			&originalAmount,
			&item.Enabled,
			&item.SortOrder,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if originalAmount.Valid {
			value := originalAmount.Float64
			item.OriginalAmount = &value
		}
		prices = append(prices, item)
	}
	return prices, rows.Err()
}

func (r *commerceCatalogRepository) selectProductGrants(ctx context.Context, productIDs []int64) ([]service.CommerceProductGroupBinding, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, product_id, group_id, grant_type, validity_days, priority, created_at, updated_at
		FROM commerce_product_group_bindings
		WHERE deleted_at IS NULL AND product_id = ANY($1)
		ORDER BY product_id ASC, priority ASC, id ASC
	`, pq.Array(productIDs))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	grants := make([]service.CommerceProductGroupBinding, 0)
	for rows.Next() {
		var item service.CommerceProductGroupBinding
		if err := rows.Scan(
			&item.ID,
			&item.ProductID,
			&item.GroupID,
			&item.GrantType,
			&item.ValidityDays,
			&item.Priority,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		grants = append(grants, item)
	}
	return grants, rows.Err()
}

func (r *commerceCatalogRepository) selectModelListings(
	ctx context.Context,
	ids []int64,
	includeInactive bool,
) ([]service.CommerceModelListing, []service.CommerceModelProductBinding, error) {
	query := `
		SELECT id, model_key, display_name, description, vendor, icon, tags, status, sort_order, recommended, metadata, created_at, updated_at
		FROM commerce_model_listings
		WHERE deleted_at IS NULL
	`
	args := make([]any, 0, 2)
	if !includeInactive {
		query += ` AND status = $1`
		args = append(args, service.CommerceProductStatusActive)
	}
	if len(ids) > 0 {
		args = append(args, pq.Array(ids))
		query += fmt.Sprintf(` AND id = ANY($%d)`, len(args))
	}
	query += ` ORDER BY sort_order ASC, id ASC`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()

	models := make([]service.CommerceModelListing, 0)
	modelIDs := make([]int64, 0)
	for rows.Next() {
		var (
			item        service.CommerceModelListing
			tagsRaw     []byte
			metadataRaw []byte
		)
		if err := rows.Scan(
			&item.ID,
			&item.ModelKey,
			&item.DisplayName,
			&item.Description,
			&item.Vendor,
			&item.Icon,
			&tagsRaw,
			&item.Status,
			&item.SortOrder,
			&item.Recommended,
			&metadataRaw,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, nil, err
		}
		item.Tags = decodeStringSliceJSON(tagsRaw)
		item.Metadata = decodeMapJSON(metadataRaw)
		models = append(models, item)
		modelIDs = append(modelIDs, item.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	if len(modelIDs) == 0 {
		return models, []service.CommerceModelProductBinding{}, nil
	}

	bindings, err := r.selectModelBindings(ctx, modelIDs)
	if err != nil {
		return nil, nil, err
	}
	return models, bindings, nil
}

func (r *commerceCatalogRepository) selectModelBindings(ctx context.Context, modelIDs []int64) ([]service.CommerceModelProductBinding, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, model_listing_id, product_id, binding_type, created_at, updated_at
		FROM commerce_model_product_bindings
		WHERE deleted_at IS NULL AND model_listing_id = ANY($1)
		ORDER BY model_listing_id ASC, id ASC
	`, pq.Array(modelIDs))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	bindings := make([]service.CommerceModelProductBinding, 0)
	for rows.Next() {
		var item service.CommerceModelProductBinding
		if err := rows.Scan(
			&item.ID,
			&item.ModelListingID,
			&item.ProductID,
			&item.BindingType,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		bindings = append(bindings, item)
	}
	return bindings, rows.Err()
}

func upsertCommerceProductRow(ctx context.Context, tx *sql.Tx, id *int64, input *service.CommerceProductUpsertInput) (int64, error) {
	tagsJSON, err := json.Marshal(input.Tags)
	if err != nil {
		return 0, err
	}
	metadataJSON, err := json.Marshal(input.Metadata)
	if err != nil {
		return 0, err
	}

	var productID int64
	if id == nil {
		err = tx.QueryRowContext(ctx, `
			INSERT INTO commerce_products (
				code, name, description, product_type, status, cover_image, tags, sort_order, recommended, metadata, created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
			RETURNING id
		`,
			input.Code,
			input.Name,
			input.Description,
			input.ProductType,
			input.Status,
			input.CoverImage,
			tagsJSON,
			input.SortOrder,
			input.Recommended,
			metadataJSON,
		).Scan(&productID)
		if err != nil {
			return 0, translatePersistenceError(err, nil, service.ErrCommerceProductCodeExists)
		}
		return productID, nil
	}

	err = tx.QueryRowContext(ctx, `
		UPDATE commerce_products
		SET code = $2,
			name = $3,
			description = $4,
			product_type = $5,
			status = $6,
			cover_image = $7,
			tags = $8,
			sort_order = $9,
			recommended = $10,
			metadata = $11,
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id
	`,
		*id,
		input.Code,
		input.Name,
		input.Description,
		input.ProductType,
		input.Status,
		input.CoverImage,
		tagsJSON,
		input.SortOrder,
		input.Recommended,
		metadataJSON,
	).Scan(&productID)
	if err != nil {
		return 0, translatePersistenceError(err, service.ErrCommerceProductNotFound, service.ErrCommerceProductCodeExists)
	}
	return productID, nil
}

func replaceCommerceProductPrices(ctx context.Context, tx *sql.Tx, productID int64, items []service.CommerceProductPriceInput) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM commerce_product_prices WHERE product_id = $1`, productID); err != nil {
		return err
	}
	for _, item := range items {
		enabled := true
		if item.Enabled != nil {
			enabled = *item.Enabled
		}
		var originalAmount any
		if item.OriginalAmount != nil {
			originalAmount = *item.OriginalAmount
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO commerce_product_prices (
				product_id, price_type, amount, currency, original_amount, enabled, sort_order, created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		`,
			productID,
			item.PriceType,
			item.Amount,
			item.Currency,
			originalAmount,
			enabled,
			item.SortOrder,
		); err != nil {
			return err
		}
	}
	return nil
}

func replaceCommerceProductGroupBindings(ctx context.Context, tx *sql.Tx, productID int64, items []service.CommerceProductGrantInput) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM commerce_product_group_bindings WHERE product_id = $1`, productID); err != nil {
		return err
	}
	for _, item := range items {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO commerce_product_group_bindings (
				product_id, group_id, grant_type, validity_days, priority, created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		`,
			productID,
			item.GroupID,
			item.GrantType,
			item.ValidityDays,
			item.Priority,
		); err != nil {
			return err
		}
	}
	return nil
}

func upsertCommerceModelRow(ctx context.Context, tx *sql.Tx, id *int64, input *service.CommerceModelListingUpsertInput) (int64, error) {
	tagsJSON, err := json.Marshal(input.Tags)
	if err != nil {
		return 0, err
	}
	metadataJSON, err := json.Marshal(input.Metadata)
	if err != nil {
		return 0, err
	}

	var modelID int64
	if id == nil {
		err = tx.QueryRowContext(ctx, `
			INSERT INTO commerce_model_listings (
				model_key, display_name, description, vendor, icon, tags, status, sort_order, recommended, metadata, created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
			RETURNING id
		`,
			input.ModelKey,
			input.DisplayName,
			input.Description,
			input.Vendor,
			input.Icon,
			tagsJSON,
			input.Status,
			input.SortOrder,
			input.Recommended,
			metadataJSON,
		).Scan(&modelID)
		if err != nil {
			return 0, translatePersistenceError(err, nil, service.ErrCommerceModelKeyExists)
		}
		return modelID, nil
	}

	err = tx.QueryRowContext(ctx, `
		UPDATE commerce_model_listings
		SET model_key = $2,
			display_name = $3,
			description = $4,
			vendor = $5,
			icon = $6,
			tags = $7,
			status = $8,
			sort_order = $9,
			recommended = $10,
			metadata = $11,
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id
	`,
		*id,
		input.ModelKey,
		input.DisplayName,
		input.Description,
		input.Vendor,
		input.Icon,
		tagsJSON,
		input.Status,
		input.SortOrder,
		input.Recommended,
		metadataJSON,
	).Scan(&modelID)
	if err != nil {
		return 0, translatePersistenceError(err, service.ErrCommerceModelNotFound, service.ErrCommerceModelKeyExists)
	}
	return modelID, nil
}

func replaceCommerceModelBindings(ctx context.Context, tx *sql.Tx, modelID int64, items []service.CommerceModelProductBindingInput) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM commerce_model_product_bindings WHERE model_listing_id = $1`, modelID); err != nil {
		return err
	}
	for _, item := range items {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO commerce_model_product_bindings (
				model_listing_id, product_id, binding_type, created_at, updated_at
			)
			VALUES ($1, $2, $3, NOW(), NOW())
		`,
			modelID,
			item.ProductID,
			item.BindingType,
		); err != nil {
			return err
		}
	}
	return nil
}

func decodeStringSliceJSON(raw []byte) []string {
	if len(raw) == 0 {
		return []string{}
	}
	var items []string
	if err := json.Unmarshal(raw, &items); err != nil {
		return []string{}
	}
	if items == nil {
		return []string{}
	}
	return items
}

func decodeMapJSON(raw []byte) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return map[string]any{}
	}
	if out == nil {
		return map[string]any{}
	}
	return out
}
