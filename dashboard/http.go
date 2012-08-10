// Dashboard exports metrics over HTTP.
//
// The index page is a human-viewable dashboard, with various graphs to
// aid visualization.
//
// /list lists the metrics currently registered, as a JSON array
// of arrays, each containing a metric name and its type.
//
// /metric returns a JSON representations of metrics.
// There is one required parameter in the query string:
//	name: Name of the metric
// For Distribution metrics, there are four more optional parameters:
//	samples: A boolean that returns a Distribution's samples if true.
//	begin, end, limit: These options are passed to Distribution#Samples
// if samples is true. begin and end must be encoded as RFC3339 timestamps.
//
// The return format is always a JSON object with two keys: Type and Value.
// Type's value is the same type as the metric (lowercase), or
// "distribution_samples" if the metric is a Distribution and samples is true.
// Value's value is the serialized version of the metric's snapshot,
// or an object with a array of integers and a count for Distribution samples.
//
// /all returns JSON representations for all registered metrics in a JSON
// object, where keys correspond to metric names and values correspond
// to the metric's JSON object, using the same format as /metric.
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
	tmpl, err := template.New("index.html").Parse(asset("index.html"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, h.registry)
}

func (h *HTTPServer) handlerJS(w http.ResponseWriter, r *http.Request) {
	var s string
	switch r.FormValue("n") {
	case "cubism":
		s = asset("js/cubism.v1.min.js")
	case "d3":
		s = asset("js/d3.v2.min.js")
	case "jquery":
		s = asset("js/jquery-1.8.0.min.js")
	case "masonry":
		s = asset("js/jquery.masonry.min.js")
	case "science":
		s = asset("js/science.v1.min.js")
	case "main":
		s = asset("js/main.js")
	case "graphs":
		s = asset("js/graphs.js")
	case "licenses":
		s = asset("js/LICENSES")
	default:
		w.WriteHeader(http.StatusNotFound)
	}

	fmt.Fprint(w, s)
}

func (h *HTTPServer) handlerAll(w http.ResponseWriter, r *http.Request) {
	m := make(map[string]typeValue)
	l := h.registry.ListMetrics()

	for i, metric := range l {
		m[i] = typeValueMetric(metric)
	}

	resp, err := json.Marshal(m)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "%s\n", resp)
}

func (h *HTTPServer) handlerMetric(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")

	metric := h.registry.FindS(name)
	if metric == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var resp []byte
	var err error

	d, ok := metric.(*metrics.Distribution)
	if ok && r.FormValue("samples") == "true" {
		tv := typeValueSamples(d, r.FormValue("begin"),
			r.FormValue("end"), r.FormValue("limit"))

		resp, err = json.Marshal(tv)
	} else {
		tv := typeValueMetric(metric)

		resp, err = json.Marshal(tv)
	}

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

// NewHTTPServer creates and starts a HTTP server on the specified address
// for a metrics registry.
func NewHTTPServer(r *metrics.Registry, addr string) HTTPServer {
	h := HTTPServer{registry: r}
	handler := http.NewServeMux()

	handler.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			h.handlerIndex(w, r)
		})
	handler.HandleFunc("/js",
		func(w http.ResponseWriter, r *http.Request) {
			h.handlerJS(w, r)
		})
	handler.HandleFunc("/all",
		func(w http.ResponseWriter, r *http.Request) {
			h.handlerAll(w, r)
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

	go h.ListenAndServe()

	return h
}
