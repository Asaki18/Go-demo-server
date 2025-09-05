package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go-orders-demo/internal/db"
	kaf "go-orders-demo/internal/kafka"
)

type Cache interface {
	Get(id string) (json.RawMessage, bool)
	Set(id string, raw json.RawMessage)
}

type Server struct {
	httpAddr string
	cache    Cache
	db       *db.Store
	prod     *kaf.Producer
	httpSrv  *http.Server
}

func New(addr string, cache Cache, store *db.Store, prod *kaf.Producer) *Server {
	s := &Server{httpAddr: addr, cache: cache, db: store, prod: prod}
	mux := http.NewServeMux()
	mux.HandleFunc("/ingest", s.handleIngest)
	mux.HandleFunc("/order/", s.handleGet)
	mux.HandleFunc("/", s.serveIndex)

	s.httpSrv = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return s
}

func (s *Server) Start() error { return s.httpSrv.ListenAndServe() }
func (s *Server) Stop(ctx context.Context) error { return s.httpSrv.Shutdown(ctx) }

func (s *Server) serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	p := filepath.Join("web", "index.html")
	f, err := os.ReadFile(p)
	if err != nil {
		http.Error(w, "index not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(f)
}

func (s *Server) handleIngest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var tmp map[string]any
	if err := json.Unmarshal(body, &tmp); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	id, _ := tmp["order_uid"].(string)
	if strings.TrimSpace(id) == "" {
		http.Error(w, "order_uid required", http.StatusBadRequest)
		return
	}

	if err := s.prod.Produce(r.Context(), body); err != nil {
		http.Error(w, fmt.Sprintf("kafka produce: %v", err), http.StatusBadGateway)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("queued"))
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/order/")
	if id == "" {
		http.Error(w, "id required", http.StatusBadRequest)
		return
	}
	if raw, ok := s.cache.Get(id); ok {
		w.Header().Set("Content-Type", "application/json")
		w.Write(raw)
		return
	}
	raw, err := s.db.Get(r.Context(), id)
	if errors.Is(err, db.ErrNotFound) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.cache.Set(id, raw)
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}