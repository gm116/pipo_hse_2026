package metrics

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type key struct {
	Method string
	Path   string
	Status int
}

type bucket struct {
	Count int64
	Sum   float64
}

type Registry struct {
	service string
	mu      sync.Mutex
	data    map[key]bucket
}

func New(service string) *Registry {
	return &Registry{service: service, data: make(map[key]bucket)}
}

func (r *Registry) Record(method, path string, status int, d time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	k := key{Method: method, Path: normalizePath(path), Status: status}
	item := r.data[k]
	item.Count++
	item.Sum += d.Seconds()
	r.data[k] = item
}

func (r *Registry) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, req)
		r.Record(req.Method, req.URL.Path, sw.status, time.Since(start))
	})
}

func (r *Registry) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprintln(w, "# HELP http_requests_total Total number of HTTP requests.")
		fmt.Fprintln(w, "# TYPE http_requests_total counter")
		fmt.Fprintln(w, "# HELP http_request_duration_seconds_sum Sum of request latencies.")
		fmt.Fprintln(w, "# TYPE http_request_duration_seconds_sum counter")
		fmt.Fprintln(w, "# HELP http_request_duration_seconds_count Count of request latencies.")
		fmt.Fprintln(w, "# TYPE http_request_duration_seconds_count counter")

		r.mu.Lock()
		defer r.mu.Unlock()
		for k, v := range r.data {
			labels := fmt.Sprintf("service=\"%s\",method=\"%s\",path=\"%s\",status=\"%d\"", r.service, k.Method, k.Path, k.Status)
			fmt.Fprintf(w, "http_requests_total{%s} %d\n", labels, v.Count)
			fmt.Fprintf(w, "http_request_duration_seconds_sum{%s} %.6f\n", labels, v.Sum)
			fmt.Fprintf(w, "http_request_duration_seconds_count{%s} %d\n", labels, v.Count)
		}
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (s *statusWriter) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func normalizePath(raw string) string {
	parts := strings.Split(raw, "/")
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		if _, err := strconv.ParseInt(parts[i], 10, 64); err == nil {
			parts[i] = ":id"
		}
	}
	return strings.Join(parts, "/")
}
