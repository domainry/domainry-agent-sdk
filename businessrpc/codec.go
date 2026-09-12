package businessrpc

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"unicode/utf8"
)

func decode(raw []byte, out any) error {
	if !utf8.Valid(raw) || len(bytes.TrimSpace(raw)) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return failure("bad_request")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	d.DisallowUnknownFields()
	if err := d.Decode(out); err != nil {
		return failure("bad_request")
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return failure("bad_request")
	}
	return nil
}
func read(r io.Reader, limit int) ([]byte, error) {
	b, err := io.ReadAll(io.LimitReader(r, int64(limit)+1))
	if err != nil || len(b) > limit {
		return nil, failure("unavailable")
	}
	return b, nil
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	raw, err := json.Marshal(v)
	if err != nil || len(raw) > MaxResponseBytes {
		raw = []byte(`{"error":{"class":"unavailable","code":"agent.business_host.unavailable"}}`)
		status = 503
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_, _ = w.Write(raw)
}
