package coordinator

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

type Server struct {
	router *Router
}

func NewServer(router *Router) *Server {
	return &Server{router: router}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /v1/route", s.route)
	mux.HandleFunc("GET /v1/scatter-plan", s.scatterPlan)
	return mux
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) route(w http.ResponseWriter, r *http.Request) {
	rawID := r.URL.Query().Get("cliente_id")
	customerID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("cliente_id must be an integer"))
		return
	}

	shard, err := s.router.ShardForCustomer(customerID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, struct {
		CustomerID int64  `json:"cliente_id"`
		Shard      string `json:"shard"`
	}{
		CustomerID: customerID,
		Shard:      shard.Name,
	})
}

func (s *Server) scatterPlan(w http.ResponseWriter, _ *http.Request) {
	shards := s.router.ScatterPlan()
	names := make([]string, 0, len(shards))
	for _, shard := range shards {
		names = append(names, shard.Name)
	}

	writeJSON(w, http.StatusOK, struct {
		Shards []string `json:"shards"`
	}{Shards: names})
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
