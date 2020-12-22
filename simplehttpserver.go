package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

type Options struct {
	ListenAddress string
	Folder        string
	Verbose       bool
}

var options Options

func main() {

	flag.StringVar(&options.ListenAddress, "listen", "0.0.0.0:8000", "Address:Port")
	flag.StringVar(&options.Folder, "path", ".", "Folder")
	flag.BoolVar(&options.Verbose, "v", false, "Verbose")
	flag.Parse()

	if flag.NArg() > 0 && options.Folder == "." {
		options.Folder = flag.Args()[0]
	}

	log.Printf("Serving %s on http://%s/...", options.Folder, options.ListenAddress)
	fmt.Println(http.ListenAndServe(options.ListenAddress, Log(http.FileServer(http.Dir(options.Folder)))))
}

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fullRequest, _ := httputil.DumpRequest(r, true)
		lrw := NewLoggingResponseWriter(w)
		handler.ServeHTTP(lrw, r)

		if options.Verbose {
			headers := new(bytes.Buffer)
			lrw.Header().Write(headers)
			log.Printf("\nRemote Address: %s\n%s\n%s %d %s\n%s\n%s\n", r.RemoteAddr, string(fullRequest), r.Proto, lrw.statusCode, http.StatusText(lrw.statusCode), headers.String(), string(lrw.Data))
		} else {
			log.Printf("%s \"%s %s %s\" %d %d", r.RemoteAddr, r.Method, r.URL, r.Proto, lrw.statusCode, len(lrw.Data))
		}
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	Data       []byte
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	// WriteHeader(int) is not called if our response implicitly returns 200 OK, so
	// we default to that status code.
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
