package repo

import (
	"context"
	"database/sql"
	"log"
	"time"
	"web_demoservice/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository interface {
	UpsertOrder(ctx context.Context, order model.Order) error
	GetOrderByUID(ctx context.Context, uid string) (model.Order, error)
}

type OrderRepo struct {
	pool *pgxpool.Pool
}

func NewOrderRepo(pool *pgxpool.Pool) *OrderRepo {
	return &OrderRepo{pool: pool}
}

func (r *OrderRepo) GetOrderByUID(ctx context.Context, uid string) (model.Order, error) {
	const q = `
SELECT 
  o.id,
  o.track_number, o.entry, o.locale, o.internal_signature, o.customer_id,
  o.delivery_service, o.shardkey, o.sm_id,
  o.date_created, o.oof_shard,

  d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,

  p.transaction, p.request_id, p.currency, p.provider,
  p.amount, p.payment_dt, p.bank, p.delivery_cost, p.goods_total, p.custom_fee
FROM orders o
LEFT JOIN delivery d ON d.order_id = o.id
LEFT JOIN payment  p ON p.order_id = o.id
WHERE o.order_uid = $1;
`

	var (
		o  model.Order
		id int64
		dt time.Time

		// delivery
		dName, dPhone, dZip, dCity, dAddress, dRegion, dEmail sql.NullString
		// payment
		pTransaction, pRequestID, pCurrency, pProvider, pBank       sql.NullString
		pAmount, pPaymentDt, pDeliveryCost, pGoodsTotal, pCustomFee sql.NullInt64
	)

	o.OrderUID = uid

	ctxQ, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := r.pool.QueryRow(ctxQ, q, uid).Scan(
		&id,
		&o.TrackNumber, &o.Entry, &o.Locale, &o.InternalSignature, &o.CustomerID,
		&o.DeliveryService, &o.ShardKey, &o.SmID,
		&dt, &o.OofShard,

		&dName, &dPhone, &dZip, &dCity, &dAddress, &dRegion, &dEmail,

		&pTransaction, &pRequestID, &pCurrency, &pProvider,
		&pAmount, &pPaymentDt, &pBank, &pDeliveryCost, &pGoodsTotal, &pCustomFee,
	); err != nil {
		return model.Order{}, err
	}

	o.DateCreated = dt.UTC().Format(time.RFC3339)
	o.Delivery.Name = dName.String
	o.Delivery.Phone = dPhone.String
	o.Delivery.Zip = dZip.String
	o.Delivery.City = dCity.String
	o.Delivery.Address = dAddress.String
	o.Delivery.Region = dRegion.String
	o.Delivery.Email = dEmail.String

	o.Payment.Transaction = pTransaction.String
	o.Payment.RequestID = pRequestID.String
	o.Payment.Currency = pCurrency.String
	o.Payment.Provider = pProvider.String
	o.Payment.Amount = int(pAmount.Int64)
	o.Payment.PaymentDt = pPaymentDt.Int64
	o.Payment.Bank = pBank.String
	o.Payment.DeliveryCost = int(pDeliveryCost.Int64)
	o.Payment.GoodsTotal = int(pGoodsTotal.Int64)
	o.Payment.CustomFee = int(pCustomFee.Int64)

	const qi = `
SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
FROM items
WHERE order_id = $1;
`
	rows, err := r.pool.Query(ctxQ, qi, id)
	if err != nil {
		return model.Order{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var it model.Item
		if err := rows.Scan(
			&it.ChrtID, &it.TrackNUmber, &it.Price, &it.RID, &it.Name, &it.Sale, &it.Size,
			&it.TotalPrice, &it.NmID, &it.Brand, &it.Status,
		); err != nil {
			return model.Order{}, err
		}
		o.Items = append(o.Items, it)
	}
	if err := rows.Err(); err != nil {
		return model.Order{}, err
	}
	return o, nil
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
