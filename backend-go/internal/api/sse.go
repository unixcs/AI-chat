package api

import (
	"net/http"
)

// sseHeaders sets the SSE response headers and returns the flusher.
func sseHeaders(w http.ResponseWriter) (http.Flusher, bool) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, false
	}
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()
	return flusher, true
}

func sseWrite(flusher http.Flusher, w http.ResponseWriter, payload string) {
	_, _ = w.Write([]byte("data: " + payload + "\n\n"))
	flusher.Flush()
}

func sseRaw(flusher http.Flusher, w http.ResponseWriter, raw string) {
	_, _ = w.Write([]byte(raw))
	flusher.Flush()
}
