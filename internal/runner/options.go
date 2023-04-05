package runner

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/projectdiscovery/goflags"
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
}

// ParseOptions parses the command line options for application
func ParseOptions() *Options {
	options := &Options{}
	flagSet := goflags.NewFlagSet()
	flagSet.SetDescription(`SimpleHTTPserver is a go enhanced version of the well known python simplehttpserver with in addition a fully customizable TCP server, both supporting TLS`)

	currentPath := "."
	if p, err := os.Getwd(); err == nil {
		currentPath = p
	}

	flagSet.CreateGroup("input", "Input",
		flagSet.StringVar(&options.ListenAddress, "listen", "0.0.0.0:8000", "Address:Port"),
	)

	flagSet.CreateGroup("config", "Config",
		flagSet.BoolVar(&options.EnableTCP, "tcp", false, "TCP Server"),
		flagSet.BoolVar(&options.TCPWithTLS, "tls", false, "Enable TCP TLS"),
		flagSet.StringVar(&options.RulesFile, "rules", "", "Rules yaml file"),
		flagSet.StringVar(&options.Folder, "path", currentPath, "Folder"),
		flagSet.BoolVar(&options.EnableUpload, "upload", false, "Enable upload via PUT"),
		flagSet.BoolVar(&options.HTTPS, "https", false, "HTTPS"),
		flagSet.StringVar(&options.TLSCertificate, "cert", "", "HTTPS Certificate"),
		flagSet.StringVar(&options.TLSKey, "key", "", "HTTPS Certificate Key"),
		flagSet.StringVar(&options.TLSDomain, "domain", "local.host", "Domain"),
		flagSet.StringVar(&options.BasicAuth, "basic-auth", "", "Basic auth (username:password),"),
		flagSet.StringVar(&options.Realm, "realm", "Please enter username and password", "Realm"),
		flagSet.BoolVar(&options.Silent, "silent", false, "Show only results in the output"),
		flagSet.BoolVar(&options.Sandbox, "sandbox", false, "Enable sandbox mode"),
		flagSet.BoolVar(&options.HTTP1Only, "http1", false, "Enable only HTTP1"),
		flagSet.IntVar(&options.MaxFileSize, "max-file-size", 50, "Max Upload File Size"),
		flagSet.IntVar(&options.MaxDumpBodySize, "max-dump-body-size", -1, "Max Dump Body Size"),
		flagSet.BoolVar(&options.Python, "py", false, "Emulate Python Style"),
		flagSet.BoolVar(&options.CORS, "cors", false, "Enable Cross-Origin Resource Sharing (CORS)"),
		flagSet.Var(&options.HTTPHeaders, "header", "Add HTTP Response Header (name: value), can be used multiple times"),
	)

	flagSet.CreateGroup("debug", "Debug",
		flagSet.BoolVar(&options.Version, "version", false, "Show version of the software"),
		flagSet.BoolVar(&options.Verbose, "verbose", false, "Verbose"),
	)

	if err := flagSet.Parse(); err != nil {
		gologger.Fatal().Msgf("%v: parse error", err.Error())
	}

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
