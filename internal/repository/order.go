package repository

import (
	"context"
	"errors"
	"fmt"
	"web_demoservice/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewOrderPostgresRepository(db *pgxpool.Pool) *OrderPostgresRepository {
	return &OrderPostgresRepository{
		db: db,
	}
}

type OrderPostgresRepository struct {
	db *pgxpool.Pool
}

func (r *OrderPostgresRepository) Create(ctx context.Context, order domain.OrderWithInformation) (err error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	// 1. Вставка основного заказа
	const qCreateOrder = `
		INSERT INTO orders.orders 
		    (order_id, track_number, entry, locale, internal_signature, customer_id, 
		     delivery_service, shardkey, sm_id, date_created, oof_shard) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (order_id) DO NOTHING;
	`
	_, err = tx.Exec(ctx, qCreateOrder,
		order.ID, order.TrackNumber, order.Entry, order.Locale,
		order.InternalSignature, order.CustomerID, order.DeliveryService,
		order.ShardKey, order.SmID, order.DateCreated, order.OofShard,
	)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	// 2. Вставка данных о доставке
	const qCreateDelivery = `
		INSERT INTO orders.delivery 
		    (order_id, name, phone, zip, city, address, region, email) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (order_id) DO NOTHING;
	`
	_, err = tx.Exec(ctx, qCreateDelivery,
		order.ID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email,
	)
	if err != nil {
		return fmt.Errorf("insert delivery: %w", err)
	}

	// 3. Обработка банка (Получаем ID по имени или создаем новый)
	var bankID int64
	err = tx.QueryRow(ctx, "SELECT id FROM banks.banks WHERE name = $1", order.Payment.Bank.Name).Scan(&bankID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = tx.QueryRow(ctx, "INSERT INTO banks.banks (name) VALUES ($1) RETURNING id", order.Payment.Bank.Name).Scan(&bankID)
			if err != nil {
				return fmt.Errorf("insert bank: %w", err)
			}
		} else {
			return fmt.Errorf("query bank: %w", err)
		}
	}

	// 4. Вставка платежа
	const qCreatePayment = `
		INSERT INTO orders.payments 
		    (order_id, transaction, request_id, currency, provider, amount, 
		     payment_dt, bank_id, delivery_cost, goods_total, custom_fee) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (transaction) DO NOTHING;
	`
	_, err = tx.Exec(ctx, qCreatePayment,
		order.ID, order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency,
		order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDt, bankID,
		order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee,
	)
	if err != nil {
		return fmt.Errorf("insert payment: %w", err)
	}

	// 5. Вставка товаров и связей
	const qCreateItem = `
		INSERT INTO orders.items 
		    (chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (rid) DO UPDATE SET status = EXCLUDED.status
		RETURNING id;
	`
	const qCreateOrderItem = `
		INSERT INTO orders.order_items (order_id, item_id) 
		VALUES ($1, $2)
		ON CONFLICT (order_id, item_id) DO NOTHING;
	`

	for _, item := range order.Items {
		var itemID int64
		err = tx.QueryRow(ctx, qCreateItem,
			item.ChrtID, item.TrackNumber, item.Price, item.RID, item.Name,
			item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status,
		).Scan(&itemID)
		if err != nil {
			return fmt.Errorf("insert item %s: %w", item.RID, err)
		}

		_, err = tx.Exec(ctx, qCreateOrderItem, order.ID, itemID)
		if err != nil {
			return fmt.Errorf("link item to order: %w", err)
		}
	}

	return nil
}

func (r *OrderPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.OrderWithInformation, error) {
	// 1. Получаем основные данные заказа
	const qGetOrder = `
		SELECT order_id, track_number, entry, locale, internal_signature, customer_id, 
		       delivery_service, shardkey, sm_id, date_created, oof_shard 
		FROM orders.orders 
		WHERE order_id = $1
	`
	var order domain.OrderWithInformation
	err := r.db.QueryRow(ctx, qGetOrder, id).Scan(
		&order.ID, &order.TrackNumber, &order.Entry, &order.Locale,
		&order.InternalSignature, &order.CustomerID, &order.DeliveryService,
		&order.ShardKey, &order.SmID, &order.DateCreated, &order.OofShard,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("order not found: %w", err)
		}
		return nil, fmt.Errorf("query order: %w", err)
	}

	// 2. Получаем данные о доставке
	const qGetDelivery = `
		SELECT name, phone, zip, city, address, region, email 
		FROM orders.delivery 
		WHERE order_id = $1
	`
	err = r.db.QueryRow(ctx, qGetDelivery, id).Scan(
		&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
		&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email,
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("query delivery: %w", err)
	}

	// 3. Получаем данные о платеже и банке
	const qGetPayment = `
		SELECT p.transaction, p.request_id, p.currency, p.provider, p.amount, 
		       p.payment_dt, p.delivery_cost, p.goods_total, p.custom_fee, 
		       b.id, b.name 
		FROM orders.payments p
		JOIN banks.banks b ON p.bank_id = b.id
		WHERE p.order_id = $1
	`
	err = r.db.QueryRow(ctx, qGetPayment, id).Scan(
		&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency,
		&order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDt,
		&order.Payment.DeliveryCost, &order.Payment.GoodsTotal, &order.Payment.CustomFee,
		&order.Payment.Bank.ID, &order.Payment.Bank.Name,
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("query payment: %w", err)
	}

	// 4. Получаем список товаров через связующую таблицу order_items
	const qGetItems = `
		SELECT i.chrt_id, i.track_number, i.price, i.rid, i.name, 
		       i.sale, i.size, i.total_price, i.nm_id, i.brand, i.status
		FROM orders.items i
		JOIN orders.order_items oi ON i.id = oi.item_id
		WHERE oi.order_id = $1
	`
	rows, err := r.db.Query(ctx, qGetItems, id)
	if err != nil {
		return nil, fmt.Errorf("query items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.Item
		err := rows.Scan(
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID, &item.Name,
			&item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		order.Items = append(order.Items, item)
	}

	return &order, nil
}

func (r *OrderPostgresRepository) GetAllLast24Hours(ctx context.Context) ([]domain.OrderWithInformation, error) {
	// 1. Получаем список ID заказов за последние 24 часа
	const qGetIDs = `
		SELECT order_id 
		FROM orders.orders 
		WHERE date_created >= NOW() - INTERVAL '24 hours'
	`
	rows, err := r.db.Query(ctx, qGetIDs)
	if err != nil {
		return nil, fmt.Errorf("query order ids: %w", err)
	}
	defer rows.Close()

	var orderIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan order id: %w", err)
		}
		orderIDs = append(orderIDs, id)
	}

	// 2. Для каждого ID вызываем GetByID (используем существующую логику сборки заказа)
	orders := make([]domain.OrderWithInformation, 0, len(orderIDs))
	for _, id := range orderIDs {
		order, err := r.GetByID(ctx, id)
		if err != nil {
			// Если один заказ не собрался, логируем и идем дальше
			continue
		}
		orders = append(orders, *order)
	}

	return orders, nil
}
