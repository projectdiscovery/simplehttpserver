package httpserver

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"

	"github.com/projectdiscovery/sslcert"
)

// Options of the http server
type Options struct {
	Folder            string
	EnableUpload      bool
	ListenAddress     string
	TLS               bool
	Certificate       string
	CertificateKey    string
	CertificateDomain string
	BasicAuthUsername string
	BasicAuthPassword string
	BasicAuthReal     string
	Verbose           bool
	Sandbox           bool
	MaxFileSize       int // 50Mb
}

// HTTPServer instance
type HTTPServer struct {
	options *Options
	layers  http.Handler
}

// LayerHandler is the interface of all layer funcs
type Middleware func(http.Handler) http.Handler

// New http server instance with options
func New(options *Options) (*HTTPServer, error) {
	var h HTTPServer
	EnableUpload = options.EnableUpload
	EnableVerbose = options.Verbose
	folder, err := filepath.Abs(options.Folder)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		return nil, errors.New("path does not exist")
	}
	options.Folder = folder
	var dir http.FileSystem
	dir = http.Dir(options.Folder)
	if options.Sandbox {
		dir = SandboxFileSystem{fs: http.Dir(options.Folder), RootFolder: options.Folder}
	}

	httpHandler := http.FileServer(dir)
	addHandler := func(newHandler Middleware) {
		httpHandler = newHandler(httpHandler)
	}

	// middleware
	if options.EnableUpload {
		addHandler(h.uploadlayer)
	}

	if options.BasicAuthUsername != "" || options.BasicAuthPassword != "" {
		addHandler(h.basicauthlayer)
	}

	httpHandler = h.loglayer(httpHandler)

	// add handler
	h.layers = httpHandler
	h.options = options

	return &h, nil
}

// ListenAndServe requests over http
func (t *HTTPServer) ListenAndServe() error {
	return http.ListenAndServe(t.options.ListenAddress, t.layers)
}

// ListenAndServeTLS requests over https
func (t *HTTPServer) ListenAndServeTLS() error {
	if t.options.Certificate == "" || t.options.CertificateKey == "" {
		tlsOptions := sslcert.DefaultOptions
		tlsOptions.Host = t.options.CertificateDomain
		tlsConfig, err := sslcert.NewTLSConfig(tlsOptions)
		if err != nil {
			return err
		}
		httpServer := &http.Server{
			Addr:      t.options.ListenAddress,
			TLSConfig: tlsConfig,
		}
		httpServer.Handler = t.layers
		return httpServer.ListenAndServeTLS("", "")
	}
	return http.ListenAndServeTLS(t.options.ListenAddress, t.options.Certificate, t.options.CertificateKey, t.layers)
}

// Close the service
func (t *HTTPServer) Close() error {
	return nil
}
