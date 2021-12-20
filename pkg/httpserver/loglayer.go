package httpserver

import (
	"bytes"
	"net/http"
	"net/http/httputil"

	"github.com/projectdiscovery/gologger"
)

// Convenience globals
var (
	EnableUpload  bool
	EnableVerbose bool
)

func (t *HTTPServer) shouldDumpBody(bodysize int64) bool {
	return t.options.MaxDumpBodySize > 0 && bodysize > t.options.MaxDumpBodySize
}

func (t *HTTPServer) loglayer(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var fullRequest []byte
		if t.shouldDumpBody(r.ContentLength) {
			fullRequest, _ = httputil.DumpRequest(r, false)
		} else {
			fullRequest, _ = httputil.DumpRequest(r, true)
		}
		lrw := newLoggingResponseWriter(w, t.options.MaxDumpBodySize)
		handler.ServeHTTP(lrw, r)

		if EnableVerbose {
			headers := new(bytes.Buffer)
			lrw.Header().Write(headers) //nolint
			gologger.Print().Msgf("\nRemote Address: %s\n%s\n%s %d %s\n%s\n%s\n", r.RemoteAddr, string(fullRequest), r.Proto, lrw.statusCode, http.StatusText(lrw.statusCode), headers.String(), string(lrw.Data))
		} else {
			gologger.Print().Msgf("%s \"%s %s %s\" %d %d", r.RemoteAddr, r.Method, r.URL, r.Proto, lrw.statusCode, lrw.Size)
		}
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode  int
	Data        []byte
	Size        int
	MaxDumpSize int64
}

func newLoggingResponseWriter(w http.ResponseWriter, maxSize int64) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK, []byte{}, 0, maxSize}
}

// Write the data
func (lrw *loggingResponseWriter) Write(data []byte) (int, error) {
	if len(lrw.Data) < int(lrw.MaxDumpSize) {
		lrw.Data = append(lrw.Data, data...)
	}
	lrw.Size += len(data)
	return lrw.ResponseWriter.Write(data)
}

// Header of the response
func (lrw *loggingResponseWriter) Header() http.Header {
	return lrw.ResponseWriter.Header()
}

// WriteHeader status code
func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
