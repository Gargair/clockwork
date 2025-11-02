package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

// decodeJSON decodes a JSON payload into dst using a strict decoder that
// disallows unknown fields and multiple JSON values.
func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return err
	}

	// Ensure there's no trailing data
	if err := dec.Decode(new(struct{})); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("invalid JSON: multiple values")
		}
		return err
	}
	return nil
}

// writeJSON writes v as a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// Best-effort write; handlers can ignore encoding errors at this layer
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes a structured ErrorResponse with the provided status, code and message.
// It includes the request ID obtained from the request context.
func writeError(w http.ResponseWriter, r *http.Request, status int, code, msg string) {
	reqID := middleware.GetReqID(r.Context())
	resp := ErrorResponse{Code: code, Message: msg, RequestID: reqID}
	writeJSON(w, status, resp)
}

// parseUUID parses a UUID from its string representation.
func parseUUID(str string) (uuid.UUID, error) {
	return uuid.Parse(str)
}

// parseTimeRFC3339 parses a time in RFC3339 format.
func parseTimeRFC3339(str string) (time.Time, error) {
	return time.Parse(time.RFC3339, str)
}
