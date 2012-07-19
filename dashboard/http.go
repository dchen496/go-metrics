package dashboard

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"metrics"
	"net/http"
	"os"
	"time"
)

type HTTPServer struct {
	registry *metrics.Registry
	*http.Server
}

func (h *HTTPServer) handlerIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, h.registry)
}

func (h *HTTPServer) handlerIndexJS(w http.ResponseWriter, r *http.Request) {
	f, _ := os.Open("")
	io.Copy(w, f)
	f.Close()
}

func (h *HTTPServer) handlerMetric(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	options := r.FormValue("options")

	metric := h.registry.FindS(name)
	if metric == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	j := JSONProcessor{}
	var opt interface{}
	var err error
	if options != "" {
		opt, err = j.decodeOption(options, metric)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	resp := metric.Process(j, name, opt)
	fmt.Fprintf(w, "%s", resp)
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

	handler.HandleFunc("/index", func(w http.ResponseWriter, r *http.Request) {
		h.handlerIndex(w, r)
	})
	handler.HandleFunc("/index.js", func(w http.ResponseWriter, r *http.Request) {
		h.handlerIndexJS(w, r)
	})
	handler.HandleFunc("/metric", func(w http.ResponseWriter, r *http.Request) {
		h.handlerMetric(w, r)
	})
	handler.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
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
