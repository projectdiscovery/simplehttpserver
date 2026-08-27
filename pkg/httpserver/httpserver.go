package httpserver

import (
	"crypto/tls"
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
	LogUA             bool // enable logging user agent
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
	HTTP1Only         bool
	MaxFileSize       int // 50Mb
	MaxDumpBodySize   int64
	Python            bool
	CORS              bool
	HTTPHeaders       []HTTPHeader
	JSONLogFile       string
}

// HTTPServer instance
type HTTPServer struct {
	options    *Options
	layers     http.Handler
	jsonLogger *JSONLogger
}

// LayerHandler is the interface of all layer funcs
type Middleware func(http.Handler) http.Handler

// New http server instance with options
func New(options *Options) (*HTTPServer, error) {
	var h HTTPServer
	EnableUpload = options.EnableUpload
	EnableVerbose = options.Verbose
	EnableLogUA = options.LogUA

	// Initialize JSON logger if specified
	if options.JSONLogFile != "" {
		jsonLogger, err := NewJSONLogger(options.JSONLogFile)
		if err != nil {
			return nil, err
		}
		h.jsonLogger = jsonLogger
	}
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

	var httpHandler http.Handler
	if options.Python {
		httpHandler = PythonStyle(dir.(http.Dir))
	} else {
		httpHandler = http.FileServer(dir)
	}

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

	if options.CORS {
		addHandler(h.corslayer)
	}

	// Wrap with router to handle /endless endpoint before applying middleware
	var router http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/endless" {
			h.endlessHandler(w, r)
			return
		}
		httpHandler.ServeHTTP(w, r)
	})

	// Apply middleware to router
	router = h.loglayer(router)
	router = h.headerlayer(router, options.HTTPHeaders)

	// add handler
	h.layers = router
	h.options = options

	return &h, nil
}

func (t *HTTPServer) makeHTTPServer(tlsConfig *tls.Config) *http.Server {
	httpServer := &http.Server{Addr: t.options.ListenAddress}
	if t.options.HTTP1Only {
		httpServer.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler))
	}
	httpServer.TLSConfig = tlsConfig
	httpServer.Handler = t.layers
	return httpServer
}

// ListenAndServe requests over http
func (t *HTTPServer) ListenAndServe() error {
	httpServer := t.makeHTTPServer(nil)
	return httpServer.ListenAndServe()
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
		httpServer := t.makeHTTPServer(tlsConfig)
		return httpServer.ListenAndServeTLS("", "")
	}
	return http.ListenAndServeTLS(t.options.ListenAddress, t.options.Certificate, t.options.CertificateKey, t.layers)
}

// endlessHandler streams endless data to the client
func (t *HTTPServer) endlessHandler(w http.ResponseWriter, r *http.Request) {
	// Set headers for streaming
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	// Flush headers immediately
	flusher, hasFlusher := w.(http.Flusher)
	if hasFlusher {
		flusher.Flush()
	}

	// 100MB in bytes
	const dataSize = 100 * 1024 * 1024

	// Generate 100MB data chunk once
	dataChunk := make([]byte, dataSize)
	pattern := []byte("Data chunk - streaming endless data...\n")
	offset := 0
	for offset < dataSize {
		copySize := len(pattern)
		if offset+copySize > dataSize {
			copySize = dataSize - offset
		}
		copy(dataChunk[offset:], pattern[:copySize])
		offset += copySize
	}

	// Stream data continuously - write and flush the 100MB chunk repeatedly
	for {
		// Check if client disconnected
		select {
		case <-r.Context().Done():
			return
		default:
		}

		// Write the 100MB chunk
		_, err := w.Write(dataChunk)
		if err != nil {
			return
		}

		// Flush after each write
		if hasFlusher {
			flusher.Flush()
		}
	}
}

// Close the service
func (t *HTTPServer) Close() error {
	if t.jsonLogger != nil {
		return t.jsonLogger.Close()
	}
	return nil
}
