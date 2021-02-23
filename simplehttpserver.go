package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"path"

	"github.com/projectdiscovery/gologger"
)

type options struct {
	ListenAddress string
	Folder        string
	Username      string
	Password      string
	Realm         string
	Certificate   string
	Key           string
	HTTPS         bool
	Verbose       bool
	Upload        bool
}

var opts options

func main() {
	flag.StringVar(&opts.ListenAddress, "listen", "0.0.0.0:8000", "Address:Port")
	flag.StringVar(&opts.Folder, "path", ".", "Folder")
	flag.BoolVar(&opts.Upload, "upload", false, "Enable upload via PUT")
	flag.BoolVar(&opts.HTTPS, "https", false, "HTTPS")
	flag.StringVar(&opts.Certificate, "cert", "", "Certificate")
	flag.StringVar(&opts.Key, "key", "", "Key")
	flag.BoolVar(&opts.Verbose, "v", false, "Verbose")
	flag.StringVar(&opts.Username, "username", "", "Basic auth username")
	flag.StringVar(&opts.Password, "password", "", "Basic auth password")
	flag.StringVar(&opts.Realm, "realm", "Please enter username and password", "Realm")

	flag.Parse()

	if flag.NArg() > 0 && opts.Folder == "." {
		opts.Folder = flag.Args()[0]
	}

	gologger.Print().Msgf("Serving %s on http://%s/...", opts.Folder, opts.ListenAddress)
	layers := loglayer(http.FileServer(http.Dir(opts.Folder)))
	if opts.Username != "" || opts.Password != "" {
		layers = loglayer(basicauthlayer(http.FileServer(http.Dir(opts.Folder))))
	}

	if opts.Upload {
		gologger.Print().Msgf("Upload enabled")
	}
	if opts.HTTPS {
		if opts.Certificate == "" || opts.Key == "" {
			gologger.Fatal().Msgf("Certificate or Key file not specified")
		}
		gologger.Print().Msgf("%s\n", http.ListenAndServeTLS(opts.ListenAddress, opts.Certificate, opts.Key, layers))
	} else {
		gologger.Print().Msgf("%s\n", http.ListenAndServe(opts.ListenAddress, layers))
	}
}

func loglayer(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fullRequest, _ := httputil.DumpRequest(r, true)
		lrw := newLoggingResponseWriter(w)
		handler.ServeHTTP(lrw, r)

		// Handles file write if enabled
		if opts.Upload && r.Method == http.MethodPut {
			data, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Println(err)
			}
			err = handleUpload(path.Base(r.URL.Path), data)
			if err != nil {
				log.Println(err)
			}
		}

		if opts.Verbose {
			headers := new(bytes.Buffer)
			lrw.Header().Write(headers) //nolint
			gologger.Print().Msgf("\nRemote Address: %s\n%s\n%s %d %s\n%s\n%s\n", r.RemoteAddr, string(fullRequest), r.Proto, lrw.statusCode, http.StatusText(lrw.statusCode), headers.String(), string(lrw.Data))
		} else {
			gologger.Print().Msgf("%s \"%s %s %s\" %d %d", r.RemoteAddr, r.Method, r.URL, r.Proto, lrw.statusCode, len(lrw.Data))
		}
	})
}

func basicauthlayer(handler http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != opts.Username || pass != opts.Password {
			w.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%s\"", opts.Realm))
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized.\n")) //nolint
			return
		}
		handler.ServeHTTP(w, r)
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

func handleUpload(file string, data []byte) error {
	return ioutil.WriteFile(file, data, 0655)
}
