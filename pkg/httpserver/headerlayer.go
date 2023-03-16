package httpserver

import (
	"net/http"
)

// HTTPHeader represents an HTTP header
type HTTPHeader struct {
	Name  string
	Value string
}

func (t *HTTPServer) headerlayer(handler http.Handler, headers []HTTPHeader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, header := range headers {
			w.Header().Set(header.Name, header.Value)
		}
		handler.ServeHTTP(w, r)
	})
}
