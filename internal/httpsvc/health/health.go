// Package health provides liveness/readiness HTTP handlers for probes.
package health

import "net/http"

// Live reports that the process is up (liveness probe).
func Live(w http.ResponseWriter, _ *http.Request) {
	writeStatus(w)
}

// Ready reports that the service can serve traffic (readiness probe). In a real
// service this would check critical dependencies (DB, caches, ...) and return
// 503 when any are unavailable.
func Ready(w http.ResponseWriter, _ *http.Request) {
	writeStatus(w)
}

func writeStatus(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
