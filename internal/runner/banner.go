package runner

import "github.com/projectdiscovery/gologger"

const banner = `
   _____ _                 __     __  __________________                                
  / ___/(_)___ ___  ____  / /__  / / / /_  __/_  __/ __ \________  ______   _____  _____
  \__ \/ / __ -__ \/ __ \/ / _ \/ /_/ / / /   / / / /_/ / ___/ _ \/ ___/ | / / _ \/ ___/
 ___/ / / / / / / / /_/ / /  __/ __  / / /   / / / ____(__  )  __/ /   | |/ /  __/ /    
/____/_/_/ /_/ /_/ .___/_/\___/_/ /_/ /_/   /_/ /_/   /____/\___/_/    |___/\___/_/     
                /_/                                                       - v0.0.6
`

// Version is the current version
const Version = `0.0.6`

// showBanner is used to show the banner to the user
func showBanner() {
	gologger.Print().Msgf("%s\n", banner)
	gologger.Print().Msgf("\t\tprojectdiscovery.io\n\n")
}
