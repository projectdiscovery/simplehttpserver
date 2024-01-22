package runner

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/gologger/levels"
	"github.com/projectdiscovery/simplehttpserver/pkg/httpserver"
)

// Options of the tool
type Options struct {
	ListenAddress   string
	Folder          string
	BasicAuth       string
	username        string
	password        string
	Realm           string
	TLSCertificate  string
	TLSKey          string
	TLSDomain       string
	HTTPS           bool
	Verbose         bool
	EnableUpload    bool
	EnableTCP       bool
	RulesFile       string
	TCPWithTLS      bool
	Version         bool
	Silent          bool
	Sandbox         bool
	MaxFileSize     int
	HTTP1Only       bool
	MaxDumpBodySize int
	Python          bool
	CORS            bool
	HTTPHeaders     HTTPHeaders
	LogRPS          bool
}

// ParseOptions parses the command line options for application
func ParseOptions() *Options {
	options := &Options{}
	flag.StringVar(&options.ListenAddress, "listen", "0.0.0.0:8000", "Address:Port")
	flag.BoolVar(&options.EnableTCP, "tcp", false, "TCP Server")
	flag.BoolVar(&options.TCPWithTLS, "tls", false, "Enable TCP TLS")
	flag.StringVar(&options.RulesFile, "rules", "", "Rules yaml file")
	currentPath := "."
	if p, err := os.Getwd(); err == nil {
		currentPath = p
	}
	flag.StringVar(&options.Folder, "path", currentPath, "Folder")
	flag.BoolVar(&options.EnableUpload, "upload", false, "Enable upload via PUT")
	flag.BoolVar(&options.HTTPS, "https", false, "HTTPS")
	flag.StringVar(&options.TLSCertificate, "cert", "", "HTTPS Certificate")
	flag.StringVar(&options.TLSKey, "key", "", "HTTPS Certificate Key")
	flag.StringVar(&options.TLSDomain, "domain", "local.host", "Domain")
	flag.BoolVar(&options.Verbose, "verbose", false, "Verbose")
	flag.StringVar(&options.BasicAuth, "basic-auth", "", "Basic auth (username:password)")
	flag.StringVar(&options.Realm, "realm", "Please enter username and password", "Realm")
	flag.BoolVar(&options.Version, "version", false, "Show version of the software")
	flag.BoolVar(&options.Silent, "silent", false, "Show only results in the output")
	flag.BoolVar(&options.Sandbox, "sandbox", false, "Enable sandbox mode")
	flag.BoolVar(&options.HTTP1Only, "http1", false, "Enable only HTTP1")
	flag.IntVar(&options.MaxFileSize, "max-file-size", 50, "Max Upload File Size")
	flag.IntVar(&options.MaxDumpBodySize, "max-dump-body-size", -1, "Max Dump Body Size")
	flag.BoolVar(&options.Python, "py", false, "Emulate Python Style")
	flag.BoolVar(&options.CORS, "cors", false, "Enable Cross-Origin Resource Sharing (CORS)")
	flag.Var(&options.HTTPHeaders, "header", "Add HTTP Response Header (name: value), can be used multiple times")
	flag.BoolVar(&options.LogRPS, "log-rps", false, "Log requests per second")
	flag.Parse()

	// Read the inputs and configure the logging
	options.configureOutput()

	showBanner()

	if options.Version {
		gologger.Info().Msgf("Current Version: %s\n", Version)
		os.Exit(0)
	}

	options.validateOptions()

	return options
}

func (options *Options) validateOptions() {
	if flag.NArg() > 0 && options.Folder == "." {
		options.Folder = flag.Args()[0]
	}

	if options.BasicAuth != "" {
		baTokens := strings.SplitN(options.BasicAuth, ":", 2)
		if len(baTokens) > 0 {
			options.username = baTokens[0]
		}
		if len(baTokens) > 1 {
			options.password = baTokens[1]
		}
	}
}

// configureOutput configures the output on the screen
func (options *Options) configureOutput() {
	// If the user desires verbose output, show verbose output
	if options.Verbose {
		gologger.DefaultLogger.SetMaxLevel(levels.LevelVerbose)
	}
	if options.Silent {
		gologger.DefaultLogger.SetMaxLevel(levels.LevelSilent)
	}
}

// FolderAbsPath of the fileserver folder
func (options *Options) FolderAbsPath() string {
	abspath, err := filepath.Abs(options.Folder)
	if err != nil {
		return options.Folder
	}
	return abspath
}

// HTTPHeaders is a slice of HTTPHeader structs
type HTTPHeaders []httpserver.HTTPHeader

func (h *HTTPHeaders) String() string {
	return fmt.Sprint(*h)
}

// Set sets a new header, which must be a string of the form 'name: value'
func (h *HTTPHeaders) Set(value string) error {
	tokens := strings.SplitN(value, ":", 2)
	if len(tokens) != 2 {
		return fmt.Errorf("header '%s' not in format 'name: value'", value)
	}

	*h = append(*h, httpserver.HTTPHeader{Name: tokens[0], Value: tokens[1]})
	return nil
}
