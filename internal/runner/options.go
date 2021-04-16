package runner

import (
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/gologger/levels"
)

type Options struct {
	ListenAddress  string
	Folder         string
	BasicAuth      string
	username       string
	password       string
	Realm          string
	TLSCertificate string
	TLSKey         string
	TLSDomain      string
	HTTPS          bool
	Verbose        bool
	EnableUpload   bool
	EnableTCP      bool
	RulesFile      string
	TCPWithTLS     bool
	Version        bool
	Silent         bool
}

// ParseOptions parses the command line options for application
func ParseOptions() *Options {
	options := &Options{}
	flag.StringVar(&options.ListenAddress, "listen", "0.0.0.0:8000", "Address:Port")
	flag.BoolVar(&options.EnableTCP, "tcp", false, "TCP Server")
	flag.BoolVar(&options.TCPWithTLS, "tls", false, "Enable TCP TLS")
	flag.StringVar(&options.RulesFile, "rules", "", "Rules yaml file")
	flag.StringVar(&options.Folder, "path", ".", "Folder")
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

func (o Options) FolderAbsPath() string {
	abspath, err := filepath.Abs(o.Folder)
	if err != nil {
		return o.Folder
	}
	return abspath
}
