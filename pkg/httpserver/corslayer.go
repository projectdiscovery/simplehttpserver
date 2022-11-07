package httpserver

import (
	"net/http"
	"strings"
)

func (t *HTTPServer) corslayer(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodOptions {
			handler.ServeHTTP(w, r)
			return
		}

		headers := w.Header()

		headers.Add("Vary", "Origin")
		headers.Add("Vary", "Access-Control-Request-Method")
		headers.Add("Vary", "Access-Control-Request-Headers")

		headers.Set("Access-Control-Allow-Origin", "*")

		reqMethod := r.Header.Get("Access-Control-Request-Method")
		if reqMethod != "" {
			headers.Set("Access-Control-Allow-Methods", strings.ToUpper(reqMethod))
		}

		w.WriteHeader(http.StatusOK)
	})
}
