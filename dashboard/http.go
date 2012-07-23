package dashboard

import (
	"encoding/json"
	"fmt"
	"html/template"
	"metrics"
	"net/http"
	"time"
)

type HTTPServer struct {
	registry *metrics.Registry
	*http.Server
}

func (h *HTTPServer) handlerIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, h.registry)
}

func (h *HTTPServer) handlerMetric(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")

	metric := h.registry.FindS(name)
	var toMarshal struct {
		Type  string
		Value interface{}
	}

	switch m := metric.(type) {
	case nil:
		w.WriteHeader(http.StatusNotFound)
		return

	case *metrics.Counter:
		toMarshal.Type = "counter"
		toMarshal.Value = m.Snapshot()

	case *metrics.Distribution:
		if r.FormValue("samples") == "true" {
			var begin, end *time.Time
			json.Unmarshal([]byte(r.FormValue("begin")), begin)
			json.Unmarshal([]byte(r.FormValue("end")), end)

			var limit uint64
			fmt.Sscanf(r.FormValue("limit"), "%d", &limit)

			toMarshal.Type = "distribution_sample"
			toMarshal.Value = m.Samples(limit, begin, end)
		} else {
			toMarshal.Type = "distribution"
			toMarshal.Value = m.Snapshot()
		}

	case *metrics.Gauge:
		snapshot := m.Snapshot()
		var stringified struct {
			Value       string
			LastUpdated time.Time
		}
		stringified.Value = snapshot.Value.String()
		stringified.LastUpdated = snapshot.LastUpdated

		toMarshal.Type = "gauge"
		toMarshal.Value = stringified
	}

	resp, err := json.MarshalIndent(toMarshal, "", "\t")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%s\n", resp)
}

func (h *HTTPServer) handlerList(w http.ResponseWriter, r *http.Request) {
	resp, err := json.Marshal(h.registry.List())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s", resp)
}

func NewHTTPServer(r *metrics.Registry, addr string) HTTPServer {
	h := HTTPServer{registry: r}
	handler := http.NewServeMux()

	handler.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			h.handlerIndex(w, r)
		})
	handler.HandleFunc("/metric",
		func(w http.ResponseWriter, r *http.Request) {
			h.handlerMetric(w, r)
		})
	handler.HandleFunc("/list",
		func(w http.ResponseWriter, r *http.Request) {
			h.handlerList(w, r)
		})
	h.Server = &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	return h
}
