package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"go-orders-demo/internal/models"
)

type Store struct {
	DB *sql.DB
}

// NewStore открывает соединение с PostgreSQL
func NewStore(connStr string) (*Store, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return &Store{DB: db}, nil
}

// SaveOrder сохраняет заказ в БД
func (s *Store) SaveOrder(ctx context.Context, o models.Order) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Вставляем заказ
	_, err = tx.ExecContext(ctx, `
		INSERT INTO orders (
			order_uid, track_number, entry, locale, internal_signature,
			customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (order_uid) DO NOTHING`,
		o.OrderUID, o.TrackNumber, o.Entry, o.Locale, o.InternalSignature,
		o.CustomerID, o.DeliveryService, o.ShardKey, o.SmID, o.DateCreated, o.OofShard,
	)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	// Вставляем доставку
	_, err = tx.ExecContext(ctx, `
		INSERT INTO deliveries (
			order_uid, name, phone, zip, city, address, region, email
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (order_uid) DO NOTHING`,
		o.OrderUID, o.Delivery.Name, o.Delivery.Phone, o.Delivery.Zip,
		o.Delivery.City, o.Delivery.Address, o.Delivery.Region, o.Delivery.Email,
	)
	if err != nil {
		return fmt.Errorf("insert delivery: %w", err)
	}

	// Вставляем оплату
	_, err = tx.ExecContext(ctx, `
		INSERT INTO payments (
			transaction, order_uid, request_id, currency, provider,
			amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (transaction) DO NOTHING`,
		o.Payment.Transaction, o.OrderUID, o.Payment.RequestID, o.Payment.Currency,
		o.Payment.Provider, o.Payment.Amount, o.Payment.PaymentDT, o.Payment.Bank,
		o.Payment.DeliveryCost, o.Payment.GoodsTotal, o.Payment.CustomFee,
	)
	if err != nil {
		return fmt.Errorf("insert payment: %w", err)
	}

	// Вставляем товары
	for _, item := range o.Items {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO items (
				chrt_id, order_uid, track_number, price, rid, name,
				sale, size, total_price, nm_id, brand, status
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
			ON CONFLICT (chrt_id) DO NOTHING`,
			item.ChrtID, o.OrderUID, item.TrackNumber, item.Price, item.Rid,
			item.Name, item.Sale, item.Size, item.TotalPrice,
			item.NmID, item.Brand, item.Status,
		)
		if err != nil {
			return fmt.Errorf("insert item: %w", err)
		}
	}

	// Коммит
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
