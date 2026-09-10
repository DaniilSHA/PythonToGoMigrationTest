package calculator

import (
	"fmt"
	"net/http"
	"strings"
)

type MetricsHandler struct {
	metrics *Metrics
}

func NewMetricsHandler(metrics *Metrics) *MetricsHandler {
	return &MetricsHandler{metrics: metrics}
}

func (h *MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	snapshot := h.metrics.snapshot()
	var body strings.Builder
	body.WriteString("# HELP calculator_http_requests_per_second POST /calc requests in each completed second; seconds_ago=1 is the most recent.\n")
	body.WriteString("# TYPE calculator_http_requests_per_second gauge\n")
	for i, count := range snapshot.rps {
		fmt.Fprintf(&body, "calculator_http_requests_per_second{seconds_ago=\"%d\"} %d\n", i+1, count)
	}
	body.WriteString("# HELP calculator_native_call_duration_seconds Exact native call duration quantiles in seconds over the last 60 seconds, excluding state lock wait time.\n")
	body.WriteString("# TYPE calculator_native_call_duration_seconds gauge\n")
	fmt.Fprintf(&body, "calculator_native_call_duration_seconds{library=\"%s\",quantile=\"0.95\"} %g\n", cLibraryKey, snapshot.cP95)
	fmt.Fprintf(&body, "calculator_native_call_duration_seconds{library=\"%s\",quantile=\"0.99\"} %g\n", cLibraryKey, snapshot.cP99)
	fmt.Fprintf(&body, "calculator_native_call_duration_seconds{library=\"%s\",quantile=\"0.95\"} %g\n", rustLibraryKey, snapshot.rustP95)
	fmt.Fprintf(&body, "calculator_native_call_duration_seconds{library=\"%s\",quantile=\"0.99\"} %g\n", rustLibraryKey, snapshot.rustP99)

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(body.String()))
}
