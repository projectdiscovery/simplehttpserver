package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

type options struct {
	ListenAddress string
	Folder        string
	Certificate   string
	Key           string
	HTTPS         bool
	Verbose       bool
}

var opts options

func main() {
	flag.StringVar(&opts.ListenAddress, "listen", "0.0.0.0:8000", "Address:Port")
	flag.StringVar(&opts.Folder, "path", ".", "Folder")
	flag.BoolVar(&opts.HTTPS, "https", false, "HTTPS")
	flag.StringVar(&opts.Certificate, "cert", "", "Certificate")
	flag.StringVar(&opts.Key, "key", "", "Key")
	flag.BoolVar(&opts.Verbose, "v", false, "Verbose")
	flag.Parse()

	if flag.NArg() > 0 && opts.Folder == "." {
		opts.Folder = flag.Args()[0]
	}

	log.Printf("Serving %s on http://%s/...", opts.Folder, opts.ListenAddress)
	if opts.HTTPS {
		if opts.Certificate == "" || opts.Key == "" {
			log.Fatal("Certificate or Key file not specified")
		}
		fmt.Println(http.ListenAndServeTLS(opts.ListenAddress, opts.Certificate, opts.Key, loglayer(http.FileServer(http.Dir(opts.Folder)))))
	} else {
		fmt.Println(http.ListenAndServe(opts.ListenAddress, loglayer(http.FileServer(http.Dir(opts.Folder)))))
	}
}

func loglayer(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fullRequest, _ := httputil.DumpRequest(r, true)
		lrw := newLoggingResponseWriter(w)
		handler.ServeHTTP(lrw, r)

		if opts.Verbose {
			headers := new(bytes.Buffer)
			lrw.Header().Write(headers) //nolint
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
