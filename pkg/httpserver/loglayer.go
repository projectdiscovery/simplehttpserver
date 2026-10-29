package httpserver

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/projectdiscovery/gologger"
)

// Convenience globals
var (
	EnableUpload  bool
	EnableVerbose bool
	EnableLogUA   bool
)

// defaultMaxDumpBodySize bounds the request body kept for logging when
// -max-dump-body-size is not set.
const defaultMaxDumpBodySize = 1024 * 1024

func (t *HTTPServer) maxDumpBodySize() int64 {
	if t.options.MaxDumpBodySize > 0 {
		return t.options.MaxDumpBodySize
	}
	return defaultMaxDumpBodySize
}

// dumpRequest dumps r with at most maxDumpBodySize bytes of its body. The
// loglayer runs before authentication, so the size kept must not depend on
// Content-Length (absent for chunked bodies). The bytes read are put back in
// front of the body, so the next handlers still read all of it.
func (t *HTTPServer) dumpRequest(r *http.Request) []byte {
	dump, _ := httputil.DumpRequest(r, false)
	if r.Body == nil || r.Body == http.NoBody {
		return dump
	}
	prefix, _ := io.ReadAll(io.LimitReader(r.Body, t.maxDumpBodySize()))
	r.Body = struct {
		io.Reader
		io.Closer
	}{io.MultiReader(bytes.NewReader(prefix), r.Body), r.Body}
	return append(dump, prefix...)
}

func (t *HTTPServer) loglayer(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// the body is read only when a log needs it
		var fullRequest []byte
		if EnableVerbose || t.jsonLogger != nil {
			fullRequest = t.dumpRequest(r)
		}
		lrw := newLoggingResponseWriter(w, t.options.MaxDumpBodySize)
		handler.ServeHTTP(lrw, r)

		// Log to JSON file if JSON logger is enabled
		if t.jsonLogger != nil {
			// Extract headers
			headers := make(map[string]string)
			for name, values := range r.Header {
				if len(values) > 0 {
					headers[name] = values[0]
				}
			}

			// Extract request body from fullRequest
			requestBody := ""
			if len(fullRequest) > 0 {
				// Find the double CRLF that separates headers from body
				bodyStart := bytes.Index(fullRequest, []byte("\r\n\r\n"))
				if bodyStart != -1 && bodyStart+4 < len(fullRequest) {
					requestBody = string(fullRequest[bodyStart+4:])
				}
			}

			// Log to JSON file
			_ = t.jsonLogger.LogRequest(r, lrw.statusCode, lrw.Size, r.UserAgent(), headers, requestBody, string(lrw.Data))
		}

		// Continue with existing console logging
		if EnableVerbose {
			headers := new(bytes.Buffer)
			lrw.Header().Write(headers) //nolint
			gologger.Print().Msgf("\n[%s]\nRemote Address: %s\n%s\n%s %d %s\n%s\n%s\n", time.Now().Format("2006-01-02 15:04:05"), r.RemoteAddr, string(fullRequest), r.Proto, lrw.statusCode, http.StatusText(lrw.statusCode), headers.String(), string(lrw.Data))
		} else {
			if EnableLogUA {
				gologger.Print().Msgf("[%s] %s \"%s %s %s\" %d %d - %s", time.Now().Format("2006-01-02 15:04:05"), r.RemoteAddr, r.Method, r.URL, r.Proto, lrw.statusCode, lrw.Size, r.UserAgent())
			} else {
				gologger.Print().Msgf("[%s] %s \"%s %s %s\" %d %d", time.Now().Format("2006-01-02 15:04:05"), r.RemoteAddr, r.Method, r.URL, r.Proto, lrw.statusCode, lrw.Size)
			}
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
