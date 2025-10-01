package httpserver

import (
	"crypto/rand"
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string, details any) {
	writeJSON(w, code, map[string]any{
		"error":   msg,
		"details": details,
	})
}

func newRefreshToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	const hex = "0123456789abcdef"
	out := make([]byte, 64)
	for i, bb := range b {
		out[i*2] = hex[bb>>4]
		out[i*2+1] = hex[bb&0x0f]
	}
	return string(out)
}
