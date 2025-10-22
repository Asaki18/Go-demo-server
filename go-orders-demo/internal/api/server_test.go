package api

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "go-orders-demo/internal/db"
)

// простой мок
type mockRepo struct {
    data map[string][]byte
}
func (m *mockRepo) SaveRaw(ctx context.Context, id string, raw json.RawMessage) error { return nil }
func (m *mockRepo) GetRaw(ctx context.Context, id string) (json.RawMessage, error) {
    v, ok := m.data[id]
    if !ok { return nil, db.ErrNotFound }
    return json.RawMessage(v), nil
}
func (m *mockRepo) LoadAllRaw(ctx context.Context, limit int) (map[string]json.RawMessage, error) { return nil, nil }
func (m *mockRepo) SaveOrder(ctx context.Context, o models.Order) error { return nil }

func TestHandleGetFromDB(t *testing.T) {
    mock := &mockRepo{data: map[string][]byte{"x":{"order_uid":"x"}}}
    c := cache.New(10)
    s := New(":0", c, mock, nil)

    req := httptest.NewRequest("GET", "/order/x", nil)
    w := httptest.NewRecorder()
    s.handleGet(w, req)
    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 got %d body=%s", w.Code, w.Body.String())
    }
}
