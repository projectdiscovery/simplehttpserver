package runner

import (
	"github.com/projectdiscovery/simplehttpserver/pkg/httpserver"
	"github.com/projectdiscovery/simplehttpserver/pkg/tcpserver"
)

// Runner is a client for running the enumeration process.
type Runner struct {
	options    *Options
	serverTCP  *tcpserver.TCPServer
	httpServer *httpserver.HTTPServer
}

func New(options *Options) (*Runner, error) {
	r := Runner{options: options}
	if r.options.EnableTCP {
		serverTCP, err := tcpserver.New(tcpserver.Options{
			Listen:  r.options.ListenAddress,
			TLS:     r.options.TCPWithTLS,
			Domain:  "local.host",
			Verbose: r.options.Verbose,
		})
		if err != nil {
			return nil, err
		}
		err = serverTCP.LoadTemplate(r.options.RulesFile)
		if err != nil {
			return nil, err
		}
		r.serverTCP = serverTCP
		return &r, nil
	}

	httpServer, err := httpserver.New(&httpserver.Options{
		Folder:            r.options.Folder,
		EnableUpload:      r.options.EnableUpload,
		ListenAddress:     r.options.ListenAddress,
		TLS:               r.options.HTTPS,
		Certificate:       r.options.TLSCertificate,
		CertificateKey:    r.options.TLSKey,
		CertificateDomain: r.options.TLSDomain,
		BasicAuthUsername: r.options.username,
		BasicAuthPassword: r.options.password,
		BasicAuthReal:     r.options.Realm,
		Verbose:           r.options.Verbose,
	})
	if err != nil {
		return nil, err
	}
	r.httpServer = httpServer

	return &r, nil
}

func (r *Runner) Run() error {
	if r.options.EnableTCP {
		return r.serverTCP.ListenAndServe()
	}

	if r.options.HTTPS {
		return r.httpServer.ListenAndServeTLS()
	}

	return r.httpServer.ListenAndServe()
}

func (r *Runner) Close() error {
	return nil
}
