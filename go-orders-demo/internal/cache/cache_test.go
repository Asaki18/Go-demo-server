package cache

import (
    "encoding/json"
    "testing"
)

func TestCacheSetGet(t *testing.T) {
    c := New(2)
    sample := json.RawMessage(`{"order_uid":"a1"}`)
    c.Set("a1", sample)
    if got, ok := c.Get("a1"); !ok {
        t.Fatal("expected key present")
    } else if string(got) != string(sample) {
        t.Fatalf("unexpected value: %s", string(got))
    }
}
