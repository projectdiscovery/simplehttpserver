package runner

import "github.com/projectdiscovery/gologger"

const banner = `
     _                 _      _     _   _                                      
 ___(_)_ __ ___  _ __ | | ___| |__ | |_| |_ _ __  ___  ___ _ ____   _____ _ __ 
/ __| | '_ ' _ \| '_ \| |/ _ \ '_ \| __| __| '_ \/ __|/ _ \ '__\ \ / / _ \ '__|
\__ \ | | | | | | |_) | |  __/ | | | |_| |_| |_) \__ \  __/ |   \ V /  __/ |   
|___/_|_| |_| |_| .__/|_|\___|_| |_|\__|\__| .__/|___/\___|_|    \_/ \___|_|   
			   |_|                        |_|                                 
`

// Version is the current version
const Version = `0.0.1`

// showBanner is used to show the banner to the user
func showBanner() {
	gologger.Print().Msgf("%s\n", banner)
	gologger.Print().Msgf("\t\tprojectdiscovery.io\n\n")

	gologger.Print().Msgf("Use with caution. You are responsible for your actions\n")
	gologger.Print().Msgf("Developers assume no liability and are not responsible for any misuse or damage.\n")
}
