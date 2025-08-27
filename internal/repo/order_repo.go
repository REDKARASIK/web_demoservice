package repo

import (
	"context"
	"log"
	"time"
	"web_demoservice/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository interface {
	UpsertOrder(ctx context.Context, order model.Order) error
}

type OrderRepo struct {
	pool *pgxpool.Pool
}

func NewOrderRepo(pool *pgxpool.Pool) *OrderRepo {
	return &OrderRepo{pool: pool}
}

func (r *OrderRepo) UpsertOrder(ctx context.Context, o model.Order) error {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()
	const qOrders = `
	INSERT INTO orders (
	               order_uid, track_number, entry, locale, internal_signature, customer_id,
	                    delivery_service, shardkey, sm_id, date_created, oof_shard
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	RETURNING id;
	`
	var t *time.Time
	if o.DateCreated != "" {
		if tt, e := time.Parse(time.RFC3339, o.DateCreated); e == nil {
			t = &tt
		}
	}
	var orderID int64
	err = tx.QueryRow(ctx, qOrders,
		o.OrderUID, o.TrackNumber, o.Entry, o.Locale, o.InternalSignature, o.CustomerID, o.DeliveryService, o.ShardKey,
		o.SmID, t, o.OofShard).Scan(&orderID)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	const qDelivery = `
	INSERT INTO delivery (
	                      order_id, name, phone, zip, city, address, region, email
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
	`
	if _, err = tx.Exec(ctx, qDelivery,
		orderID, o.Delivery.Name, o.Delivery.Phone, o.Delivery.Zip, o.Delivery.City, o.Delivery.Address,
		o.Delivery.Region, o.Delivery.Email); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	const qPayment = `
	INSERT INTO payment (
	                     order_id, transaction, request_id, currency, provider, amount, payment_dt, bank,
	                     delivery_cost, goods_total, custom_fee
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);
	`
	if _, err = tx.Exec(ctx, qPayment, orderID, o.Payment.Transaction, o.Payment.RequestID, o.Payment.Currency,
		o.Payment.Provider, o.Payment.Amount, o.Payment.PaymentDt, o.Payment.Bank, o.Payment.DeliveryCost, o.Payment.GoodsTotal, o.Payment.CustomFee); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	const qItem = `
	INSERT INTO items (
	                  order_id, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);
	`
	for _, item := range o.Items {
		if _, err := tx.Exec(ctx, qItem,
			orderID, item.ChrtID, item.TrackNUmber, item.Price, item.RID, item.Name, item.Sale, item.Size,
			item.TotalPrice, item.NmID, item.Brand, item.Status); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
	}
	return func() error {
		err := tx.Commit(ctx)
		if err != nil {
			return err
		}
		log.Println("INFO: DbPool Added and Committed")
		return nil
	}()
}
