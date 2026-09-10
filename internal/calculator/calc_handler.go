package calculator

import (
	"net/http"
	"strconv"
	"strings"
)

type CalcHandler struct {
	state   *State
	metrics *Metrics
}

func NewCalcHandler(state *State, metrics *Metrics) *CalcHandler {
	return &CalcHandler{state: state, metrics: metrics}
}

func (h *CalcHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.metrics.recordRequest()

	var rawNum string
	for _, value := range r.URL.Query()["num"] {
		if value != "" {
			rawNum = value
			break
		}
	}
	if rawNum == "" {
		h.respond(w, http.StatusBadRequest, []byte("missing 'num' query parameter"))
		return
	}

	num, err := strconv.ParseInt(strings.TrimSpace(rawNum), 10, 64)
	if err != nil {
		h.respond(w, http.StatusBadRequest, []byte("'num' must be an integer"))
		return
	}

	h.state.Add(num)

	h.respond(w, http.StatusOK, []byte("ok"))
}

func (h *CalcHandler) respond(w http.ResponseWriter, code int, body []byte) {
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(code)
	if len(body) > 0 {
		_, _ = w.Write(body)
	}
}
