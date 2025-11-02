package http

import (
	"bytes"
	"context"
	"encoding/json"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

// mountRoutes creates a router and mounts the given registrar under the base path.
func mountRoutes(base string, register func(r chi.Router)) *chi.Mux {
	r := chi.NewRouter()
	r.Route(base, register)
	return r
}

// setRouteParams attaches chi route params to the request.
func setRouteParams(req *stdhttp.Request, params map[string]string) *stdhttp.Request {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return req.WithContext(contextWithRoute(req, rctx))
}

func contextWithRoute(req *stdhttp.Request, rctx *chi.Context) context.Context {
	return context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
}

// doRequest performs an HTTP request against the router with optional JSON body and params.
func doRequest(r *chi.Mux, method, url string, body []byte, params map[string]string) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, url, reader)
	if params != nil {
		req = setRouteParams(req, params)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// mustJSON marshals a value into JSON, panicking in tests if it fails.
func mustJSON(t testing.TB, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}
	return b
}
