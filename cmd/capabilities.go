package cmd

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mrz1836/go-validate"
	"github.com/mrz1836/paymail-inspector/chalker"
	"github.com/mrz1836/paymail-inspector/paymail"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

// capabilitiesCmd represents the capabilities command
var capabilitiesCmd = &cobra.Command{
	Use:   "capabilities",
	Short: "Get the capabilities of the paymail domain",
	Long: chalk.Green.Color(`
                          ___.   .__.__  .__  __  .__               
  ____ _____  ___________ \_ |__ |__|  | |__|/  |_|__| ____   ______
_/ ___\\__  \ \____ \__  \ | __ \|  |  | |  \   __\  |/ __ \ /  ___/
\  \___ / __ \|  |_> > __ \| \_\ \  |  |_|  ||  | |  \  ___/ \___ \ 
 \___  >____  /   __(____  /___  /__|____/__||__| |__|\___  >____  >
     \/     \/|__|       \/    \/                         \/     \/`) + `
` + chalk.Yellow.Color(`
This command will return the capabilities for a given paymail domain.

Capability Discovery is the process by which a paymail client learns the supported 
features of a paymail service and their respective endpoints and configurations.

Drawing inspiration from RFC 5785 and IANA's Well-Known URIs resource, the Capability Discovery protocol 
dictates that a machine-readable document is placed in a predictable location on a web server.

Read more at: `+chalk.Cyan.Color("http://bsvalias.org/02-02-capability-discovery.html")),
	Aliases: []string{"c", "abilities", "inspect", "lookup"},
	Example: applicationName + " capabilities " + defaultDomainName,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return chalker.Error("capabilities requires either a domain or paymail address")
		} else if len(args) > 1 {
			return chalker.Error("capabilities only supports one domain or address at a time")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		// Extract the parts given
		domain, _ := paymail.ExtractParts(args[0])

		// Check for a real domain (require at least one period)
		if !strings.Contains(domain, ".") {
			chalker.Log(chalker.ERROR, fmt.Sprintf("Domain name is invalid: %s", domain))
			return
		} else if !validate.IsValidDNSName(domain) { // Basic DNS check (not a REAL domain name check)
			chalker.Log(chalker.ERROR, fmt.Sprintf("Domain name failed DNS check: %s", domain))
			return
		}

		// Get the capabilities
		capabilities, err := getCapabilities(domain, false)
		if err != nil {
			if strings.Contains(err.Error(), "context deadline exceeded") {
				chalker.Log(chalker.WARN, fmt.Sprintf("No capabilities found for: %s", domain))
			} else {
				chalker.Log(chalker.ERROR, fmt.Sprintf("Error: %s", err.Error()))
			}
			return
		}

		// Rendering profile information
		displayHeader(chalker.DEFAULT, fmt.Sprintf("Listing %d capabilities...", len(capabilities.Capabilities)))

		// Show all the found capabilities
		// todo: loop known BRFCs and display "more" info in this display for all detected BRFCs
		for key, val := range capabilities.Capabilities {
			valType := reflect.TypeOf(val).String()
			if valType == "string" {
				chalker.Log(chalker.INFO, fmt.Sprintf("%s: %-28v %s: %s", chalk.White.Color("Capability"), chalk.Cyan.Color(key), chalk.White.Color("Target"), chalk.Yellow.Color(fmt.Sprintf("%s", val))))
			} else if valType == "bool" { // See: http://bsvalias.org/04-02-sender-validation.html
				if val.(bool) {
					chalker.Log(chalker.INFO, fmt.Sprintf("%s: %-28v Is    : %s", chalk.White.Color("Capability"), chalk.Cyan.Color(key), chalk.Green.Color("Enabled")))
				} else {
					chalker.Log(chalker.INFO, fmt.Sprintf("%s: %-28v Is    : %s", chalk.White.Color("Capability"), chalk.Cyan.Color(key), chalk.Magenta.Color("Disabled")))
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(capabilitiesCmd)
}
