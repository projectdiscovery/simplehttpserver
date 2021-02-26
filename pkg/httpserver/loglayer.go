package httpserver

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"path"

	"github.com/projectdiscovery/gologger"
)

// Convenience globals
var (
	EnableUpload  bool
	EnableVerbose bool
)

func (t *HTTPServer) loglayer(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fullRequest, _ := httputil.DumpRequest(r, true)
		lrw := newLoggingResponseWriter(w)
		handler.ServeHTTP(lrw, r)

		// Handles file write if enabled
		if EnableUpload && r.Method == http.MethodPut {
			data, err := ioutil.ReadAll(r.Body)
			if err != nil {
				gologger.Print().Msgf("%s\n", err)
			}
			err = handleUpload(path.Base(r.URL.Path), data)
			if err != nil {
				gologger.Print().Msgf("%s\n", err)
			}
		}

		if EnableVerbose {
			headers := new(bytes.Buffer)
			lrw.Header().Write(headers) //nolint
			gologger.Print().Msgf("\nRemote Address: %s\n%s\n%s %d %s\n%s\n%s\n", r.RemoteAddr, string(fullRequest), r.Proto, lrw.statusCode, http.StatusText(lrw.statusCode), headers.String(), string(lrw.Data))
		} else {
			gologger.Print().Msgf("%s \"%s %s %s\" %d %d", r.RemoteAddr, r.Method, r.URL, r.Proto, lrw.statusCode, len(lrw.Data))
		}
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	Data       []byte
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK, []byte{}}
}

func (lrw *loggingResponseWriter) Write(data []byte) (int, error) {
	lrw.Data = append(lrw.Data, data...)
	return lrw.ResponseWriter.Write(data)
}

func (lrw *loggingResponseWriter) Header() http.Header {
	return lrw.ResponseWriter.Header()
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
