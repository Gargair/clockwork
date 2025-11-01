package http

import (
	"context"
	"database/sql"
	"encoding/json"
	stdhttp "net/http"
	"time"

	"github.com/Gargair/clockwork/server/internal/clock"
	"github.com/Gargair/clockwork/server/internal/db"
)

// HealthResponse is the JSON payload returned by the /healthz endpoint.
type HealthResponse struct {
	OK   bool   `json:"ok"`
	DB   string `json:"db"`
	Time string `json:"time"`
}

// HealthzHandler serves the /healthz route.
type HealthzHandler struct {
	db  *sql.DB
	clk clock.Clock
}

func (h HealthzHandler) ServeHTTP(w stdhttp.ResponseWriter, req *stdhttp.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), time.Second)
	defer cancel()

	dbStatus := "down"
	if err := db.Health(ctx, h.db); err == nil {
		dbStatus = "up"
	} else {
		w.WriteHeader(stdhttp.StatusServiceUnavailable)
	}

	resp := HealthResponse{
		OK:   dbStatus == "up",
		DB:   dbStatus,
		Time: h.clk.Now().Format(time.RFC3339Nano),
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
