package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"go-orders-demo/internal/models"
)

var ErrNotFound = errors.New("not found")

type SQLStore struct {
	db *sql.DB
}

func NewSQLStore(connStr string) (*SQLStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return &SQLStore{db: db}, nil
}

// SaveRaw - сохраняет целый JSON в orders.payload (совместимость)
func (s *SQLStore) SaveRaw(ctx context.Context, id string, raw json.RawMessage) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO orders(order_uid, payload, updated_at) VALUES ($1, $2, now())
		ON CONFLICT (order_uid) DO UPDATE SET payload = EXCLUDED.payload, updated_at = now()
	`, id, raw)
	return err
}

func (s *SQLStore) GetRaw(ctx context.Context, id string) (json.RawMessage, error) {
	var raw []byte
	err := s.db.QueryRowContext(ctx, `SELECT payload FROM orders WHERE order_uid=$1`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return json.RawMessage(raw), err
}

func (s *SQLStore) LoadAllRaw(ctx context.Context, limit int) (map[string]json.RawMessage, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT order_uid, payload FROM orders ORDER BY updated_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := make(map[string]json.RawMessage)
	for rows.Next() {
		var id string
		var raw []byte
		if err := rows.Scan(&id, &raw); err != nil {
			return nil, err
		}
		res[id] = json.RawMessage(raw)
	}
	return res, rows.Err()
}

// SaveOrder — нормализованная вставка (orders, deliveries, payments, items)
// Используем транзакцию
func (s *SQLStore) SaveOrder(ctx context.Context, o models.Order) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// insert order row (and payload as fallback)
	payload, _ := json.Marshal(o)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO orders (order_uid, payload, track_number, entry, locale, internal_signature,
			customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12, now())
		ON CONFLICT (order_uid) DO UPDATE SET payload=EXCLUDED.payload, updated_at=now()
	`, o.OrderUID, payload, o.TrackNumber, o.Entry, o.Locale, o.InternalSignature,
		o.CustomerID, o.DeliveryService, o.ShardKey, o.SmID, o.DateCreated, o.OofShard)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO deliveries (order_uid, name, phone, zip, city, address, region, email)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (order_uid) DO UPDATE SET name=EXCLUDED.name, phone=EXCLUDED.phone
	`, o.OrderUID, o.Delivery.Name, o.Delivery.Phone, o.Delivery.Zip,
		o.Delivery.City, o.Delivery.Address, o.Delivery.Region, o.Delivery.Email)
	if err != nil {
		return fmt.Errorf("insert delivery: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO payments (transaction, order_uid, request_id, currency, provider,
			amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (transaction) DO UPDATE SET amount=EXCLUDED.amount
	`, o.Payment.Transaction, o.OrderUID, o.Payment.RequestID, o.Payment.Currency,
		o.Payment.Provider, o.Payment.Amount, o.Payment.PaymentDT, o.Payment.Bank,
		o.Payment.DeliveryCost, o.Payment.GoodsTotal, o.Payment.CustomFee)
	if err != nil {
		return fmt.Errorf("insert payment: %w", err)
	}

	for _, it := range o.Items {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO items (chrt_id, order_uid, track_number, price, rid, name,
				sale, size, total_price, nm_id, brand, status)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
			ON CONFLICT (chrt_id) DO UPDATE SET price=EXCLUDED.price, total_price=EXCLUDED.total_price
		`, it.ChrtID, o.OrderUID, it.TrackNumber, it.Price, it.Rid,
			it.Name, it.Sale, it.Size, it.TotalPrice, it.NmID, it.Brand, it.Status)
		if err != nil {
			return fmt.Errorf("insert item: %w", err)
		}
	}

	// если нужно — обновляем updated_at
	_, _ = tx.ExecContext(ctx, `UPDATE orders SET updated_at = now() WHERE order_uid=$1`, o.OrderUID)

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}
