package httpserver

import (
	"net/http"

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
}

// HTTPServer instance
type HTTPServer struct {
	options *Options
	layers  http.Handler
}

// New http server instance with options
func New(options *Options) (*HTTPServer, error) {
	var h HTTPServer
	EnableUpload = options.EnableUpload
	EnableVerbose = options.Verbose
	layers := h.loglayer(http.FileServer(http.Dir(options.Folder)))
	if options.BasicAuthUsername != "" || options.BasicAuthPassword != "" {
		layers = h.loglayer(h.basicauthlayer(http.FileServer(http.Dir(options.Folder))))
	}

	return &HTTPServer{options: options, layers: layers}, nil
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
