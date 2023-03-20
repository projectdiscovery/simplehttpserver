package tcpserver

import (
	"context"
	"crypto/tls"
	"errors"
	"os"
	"net"
	"sync"
	"time"

	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/sslcert"
	"gopkg.in/yaml.v2"
)

const readTimeout = 5

// Options of the tcp server
type Options struct {
	Listen      string
	TLS         bool
	Certificate string
	Key         string
	Domain      string
	rules       []Rule
	Verbose     bool
}

// CallBackFunc handles what is send back to the client, based on the incomming question
type CallBackFunc func(ctx context.Context, question []byte) (answer []byte, err error)

// TCPServer instance
type TCPServer struct {
	options  *Options
	listener net.Listener

	// Callbacks to retrieve information about the system
	HandleMessageFnc CallBackFunc

	mux   sync.RWMutex
	rules []Rule
}

// New tcp server instance with specified options
func New(options *Options) (*TCPServer, error) {
	srv := &TCPServer{options: options}
	srv.HandleMessageFnc = srv.BuildResponseWithContext
	srv.rules = options.rules
	return srv, nil
}

// AddRule to the server
func (t *TCPServer) AddRule(rule Rule) error {
	t.mux.Lock()
	defer t.mux.Unlock()

	t.rules = append(t.rules, rule)
	return nil
}

// ListenAndServe requests
func (t *TCPServer) ListenAndServe() error {
	listener, err := net.Listen("tcp4", t.options.Listen)
	if err != nil {
		return err
	}
	t.listener = listener
	return t.run()
}

func (t *TCPServer) handleConnection(conn net.Conn, callback CallBackFunc) error {
	defer conn.Close() //nolint

	// Create Context
	ctx := context.WithValue(context.Background(), Addr, conn.RemoteAddr())

	buf := make([]byte, 4096)
	for {
		if err := conn.SetReadDeadline(time.Now().Add(readTimeout * time.Second)); err != nil {
			gologger.Info().Msgf("%s\n", err)
		}
		n, err := conn.Read(buf)
		if err != nil {
			return err
		}

		gologger.Print().Msgf("%s\n", buf[:n])

		resp, err := callback(ctx, buf[:n])
		if err != nil {
			gologger.Info().Msgf("Closing connection: %s\n", err)
			return err
		}

		if _, err := conn.Write(resp); err != nil {
			gologger.Info().Msgf("%s\n", err)
		}

		gologger.Print().Msgf("%s\n", resp)
	}
}

// ListenAndServeTLS requests over tls
func (t *TCPServer) ListenAndServeTLS() error {
	var tlsConfig *tls.Config
	if t.options.Certificate != "" && t.options.Key != "" {
		cert, err := tls.LoadX509KeyPair(t.options.Certificate, t.options.Key)
		if err != nil {
			return err
		}
		tlsConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
	} else {
		tlsOptions := sslcert.DefaultOptions
		tlsOptions.Host = t.options.Domain
		cfg, err := sslcert.NewTLSConfig(tlsOptions)
		if err != nil {
			return err
		}
		tlsConfig = cfg
	}

	listener, err := tls.Listen("tcp", t.options.Listen, tlsConfig)
	if err != nil {
		return err
	}
	t.listener = listener
	return t.run()
}

func (t *TCPServer) run() error {
	for {
		c, err := t.listener.Accept()
		if err != nil {
			return err
		}
		go t.handleConnection(c, t.HandleMessageFnc) //nolint
	}
}

// Close the service
func (t *TCPServer) Close() error {
	return t.listener.Close()
}

// LoadTemplate from yaml
func (t *TCPServer) LoadTemplate(templatePath string) error {
	var config RulesConfiguration
	yamlFile, err := os.ReadFile(templatePath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return err
	}

	t.mux.Lock()
	defer t.mux.Unlock()

	t.rules = make([]Rule, 0)
	for _, ruleTemplate := range config.Rules {
		rule, err := NewRuleFromTemplate(ruleTemplate)
		if err != nil {
			return err
		}
		t.rules = append(t.rules, *rule)
	}

	gologger.Info().Msgf("TCP configuration loaded. Rules: %d\n", len(t.rules))

	return nil
}

// MatchRule returns the rule, which was matched first
func (t *TCPServer) MatchRule(data []byte) (rule Rule, err error) {
	t.mux.RLock()
	defer t.mux.RUnlock()

	// Process all the rules
	for _, rule := range t.rules {
		if rule.MatchInput(data) {
			return rule, nil
		}
	}
	return Rule{}, errors.New("no matched rule")
}

// BuildResponseWithContext is a wrapper with context
func (t *TCPServer) BuildResponseWithContext(ctx context.Context, data []byte) ([]byte, error) {
	return t.BuildResponse(data)
}

// BuildResponseWithContext is a wrapper with context
func (t *TCPServer) BuildRuleResponse(ctx context.Context, data []byte) ([]byte, error) {
	addr := "unknown"
	if netAddr, ok := ctx.Value(Addr).(net.Addr); ok {
		addr = netAddr.String()
	}
	rule, err := t.MatchRule(data)
	if err != nil {
		return []byte(":) "), err
	}

	gologger.Info().Msgf("Incoming TCP request(%s) from: %s\n", rule.Name, addr)

	return []byte(rule.Response), nil
}
