package cmd

import (
	"fmt"
	"strings"

	"github.com/mrz1836/paymail-inspector/chalker"
	"github.com/mrz1836/paymail-inspector/paymail"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
)

const (
	defaultSatoshiValue = 1000
)

// p2pCmd represents the p2p command
var p2pCmd = &cobra.Command{
	Use:        "p2p",
	Short:      "Starts a new p2p payment request",
	Long:       `This command will start a new p2p request with the receiver.`,
	Aliases:    []string{"send"},
	SuggestFor: []string{"sending"},
	Example:    "p2p this@address.com",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return chalker.Error("%s requires a paymail address")
		} else if len(args) > 1 {
			return chalker.Error("p2p only supports one address at a time")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {

		// Set the domain and paymail
		domain, paymailAddress := paymail.ExtractParts(args[0])

		// Did we get a paymail address?
		if len(paymailAddress) == 0 {
			chalker.Log(chalker.ERROR, "paymail address not found or invalid")
			return
		}

		// Validate the paymail address and domain (error already shown)
		if ok := validatePaymailAndDomain(paymailAddress, domain); !ok {
			return
		}

		// Get the capabilities
		capabilities, err := getCapabilities(domain)
		if err != nil {
			if strings.Contains(err.Error(), "context deadline exceeded") {
				chalker.Log(chalker.WARN, fmt.Sprintf("no capabilities found for: %s", domain))
			} else {
				chalker.Log(chalker.ERROR, fmt.Sprintf("error: %s", err.Error()))
			}
			return
		}

		// Does the paymail provider have the capability?
		if len(capabilities.P2PPaymentDestination) == 0 {
			chalker.Log(chalker.ERROR, fmt.Sprintf("%s is missing a required capability: %s", domain, paymail.CapabilityP2PPaymentDestination))
			return
		}

		// Set the satoshis
		if satoshis <= 0 {
			satoshis = defaultSatoshiValue
		}

		// Get the alias of the address
		parts := strings.Split(paymailAddress, "@")

		// Fire the request
		var p2pResponse *paymail.P2PPaymentDestinationResponse
		if p2pResponse, err = paymail.GetP2PPaymentDestination(
			capabilities.P2PPaymentDestination,
			parts[0],
			domain,
			&paymail.P2PPaymentDestinationRequest{Satoshis: satoshis},
		); err != nil {
			chalker.Log(chalker.ERROR, fmt.Sprintf("p2p payment request failed: %s", err.Error()))
			return
		}

		// Attempt to get a public profile if the capability is found
		if len(capabilities.PublicProfile) > 0 && !skipPublicProfile {
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("getting public profile for: %s...", chalk.Cyan.Color(parts[0]+"@"+domain)))
			var profile *paymail.PublicProfileResponse
			if profile, err = paymail.GetPublicProfile(capabilities.PublicProfile, parts[0], domain); err != nil {
				chalker.Log(chalker.ERROR, fmt.Sprintf("get public profile failed: %s", err.Error()))
				return
			} else if profile != nil {
				if len(profile.Name) > 0 {
					chalker.Log(chalker.DEFAULT, fmt.Sprintf("name: %s", chalk.Cyan.Color(profile.Name)))
				}
				if len(profile.Avatar) > 0 {
					chalker.Log(chalker.DEFAULT, fmt.Sprintf("avatar: %s", chalk.Cyan.Color(profile.Avatar)))
				}
			}
		}

		// If there is a reference
		if len(p2pResponse.Reference) > 0 {
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("payment reference: %s", chalk.Cyan.Color(p2pResponse.Reference)))
		}

		// Output the results
		for index, output := range p2pResponse.Outputs {

			// Show output script & amount
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("output #%d: script: %s", index+1, chalk.Cyan.Color(output.Script)))
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("output #%d: satoshis: %s", index+1, chalk.Cyan.Color(fmt.Sprintf("%d", output.Satoshis))))
			chalker.Log(chalker.DEFAULT, fmt.Sprintf("output #%d: address: %s", index+1, chalk.Cyan.Color(output.Address)))
		}
	},
}

func init() {
	rootCmd.AddCommand(p2pCmd)

	// Set the amount for the sender request
	p2pCmd.Flags().Uint64Var(&satoshis, "satoshis", 0, "Amount in satoshis for the the incoming transaction(s)")
}