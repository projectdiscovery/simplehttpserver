package httpserver

import (
	"fmt"
	"net/http"
)

func (t *HTTPServer) basicauthlayer(handler http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != t.options.BasicAuthUsername || pass != t.options.BasicAuthPassword {
			w.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%s\"", t.options.BasicAuthReal))
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized.\n")) //nolint
			return
		}
		handler.ServeHTTP(w, r)
	})
}
