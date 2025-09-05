package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	_ "github.com/lib/pq"
)

var ErrNotFound = errors.New("not found")

type Store struct {
	db *sql.DB
}

func New(dsn string) (*Store, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Save(ctx context.Context, id string, raw json.RawMessage) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO orders(order_uid, payload) VALUES ($1, $2)
		ON CONFLICT (order_uid) DO UPDATE SET payload = EXCLUDED.payload, updated_at = now()
	`, id, raw)
	return err
}

func (s *Store) Get(ctx context.Context, id string) (json.RawMessage, error) {
	var raw []byte
	err := s.db.QueryRowContext(ctx, `SELECT payload FROM orders WHERE order_uid=$1`, id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return json.RawMessage(raw), err
}

func (s *Store) LoadAll(ctx context.Context, limit int) (map[string]json.RawMessage, error) {
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
