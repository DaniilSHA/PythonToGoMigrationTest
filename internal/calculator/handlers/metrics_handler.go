package handlers

import (
	"net/http"
)

type Metrics interface {
	Prometheus() string
}

type MetricsHandler struct {
	metrics Metrics
}

func NewMetricsHandler(metrics Metrics) *MetricsHandler {
	return &MetricsHandler{metrics: metrics}
}

func (h *MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body := h.metrics.Prometheus()

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(body))
}
