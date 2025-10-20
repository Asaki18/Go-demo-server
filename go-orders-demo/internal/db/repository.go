package db

import (
	"context"
	"encoding/json"
	"go-orders-demo/internal/models"
)

type Repository interface {
	// методы для кэша / API
	SaveRaw(ctx context.Context, id string, raw json.RawMessage) error
	GetRaw(ctx context.Context, id string) (json.RawMessage, error)
	LoadAllRaw(ctx context.Context, limit int) (map[string]json.RawMessage, error)

	// новый нормализованный метод
	SaveOrder(ctx context.Context, o models.Order) error
}
