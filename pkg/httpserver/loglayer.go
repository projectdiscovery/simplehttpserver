package httpserver

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"path"
	"path/filepath"

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
			// sandbox - calcolate absolute path
			if t.options.Sandbox {
				absPath, err := filepath.Abs(filepath.Join(t.options.Folder, r.URL.Path))
				if err != nil {
					gologger.Print().Msgf("%s\n", err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				// check if the path is within the configured folder
				pattern := t.options.Folder + string(filepath.Separator) + "*"
				matched, err := filepath.Match(pattern, absPath)
				if err != nil {
					gologger.Print().Msgf("%s\n", err)
					w.WriteHeader(http.StatusBadRequest)
					return
				} else if !matched {
					gologger.Print().Msg("pointing to unauthorized directory")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			}

			var (
				data []byte
				err  error
			)
			if t.options.Sandbox {
				maxFileSize := toMb(t.options.MaxFileSize)
				// check header content length
				if r.ContentLength > maxFileSize {
					gologger.Print().Msg("request too large")
					return
				}
				// body max length
				r.Body = http.MaxBytesReader(w, r.Body, maxFileSize)
			}

			data, err = ioutil.ReadAll(r.Body)
			if err != nil {
				gologger.Print().Msgf("%s\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			err = handleUpload(path.Base(r.URL.Path), data)
			if err != nil {
				gologger.Print().Msgf("%s\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
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

// Write the data
func (lrw *loggingResponseWriter) Write(data []byte) (int, error) {
	lrw.Data = append(lrw.Data, data...)
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
