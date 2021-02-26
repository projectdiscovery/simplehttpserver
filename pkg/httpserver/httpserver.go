package httpserver

import (
	"errors"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"syscall"

	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/simplehttpserver/pkg/sslcert"
)

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

type HTTPServer struct {
	options  *Options
	layers   http.Handler
	listener net.Listener
}

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

func (t *HTTPServer) ListenAndServe() error {
	var err error
retry_listen:
	gologger.Print().Msgf("Serving %s on http://%s/...", t.options.Folder, t.options.ListenAddress)
	err = http.ListenAndServe(t.options.ListenAddress, t.layers)
	if err != nil {
		if isErrorAddressAlreadyInUse(err) {
			gologger.Print().Msgf("Can't listen on %s: %s - retrying with another port\n", t.options.ListenAddress, err)
			newListenAddress, err := incPort(t.options.ListenAddress)
			if err != nil {
				return err
			}
			t.options.ListenAddress = newListenAddress
			goto retry_listen
		}
	}
	return nil
}

func (t *HTTPServer) ListenAndServeTLS() error {
	gologger.Print().Msgf("Serving %s on https://%s/...", t.options.Folder, t.options.ListenAddress)
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

func isErrorAddressAlreadyInUse(err error) bool {
	var eOsSyscall *os.SyscallError
	if !errors.As(err, &eOsSyscall) {
		return false
	}
	var errErrno syscall.Errno // doesn't need a "*" (ptr) because it's already a ptr (uintptr)
	if !errors.As(eOsSyscall, &errErrno) {
		return false
	}
	if errErrno == syscall.EADDRINUSE {
		return true
	}
	const WSAEADDRINUSE = 10048
	if runtime.GOOS == "windows" && errErrno == WSAEADDRINUSE {
		return true
	}
	return false
}

func incPort(address string) (string, error) {
	addrOrig, portOrig, err := net.SplitHostPort(address)
	if err != nil {
		return address, err
	}

	// increment port
	portNumber, err := strconv.Atoi(portOrig)
	if err != nil {
		return address, err
	}
	portNumber++
	newPort := strconv.FormatInt(int64(portNumber), 10)

	return net.JoinHostPort(addrOrig, newPort), nil
}
