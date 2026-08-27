package commands

import (
	"github.com/spf13/cobra"
)

// GET /preferences — the authenticated user's own settings (locale, theme,
// accent, time_zone, export_locale). Read-only here: the app owns the form that
// changes them.
//
// `time_zone` is the one a client cannot do without. The server files every
// entry under the day its started_at falls on IN THE USER'S ZONE, and renders
// JSON timestamps in the application's zone, which is a different thing — so a
// client that groups by the timestamp's own offset, or by the machine's clock,
// draws different day boundaries than the app for the same entries.
var preferencesCmd = &cobra.Command{
	Use:     "preferences",
	Aliases: []string{"prefs"},
	Short:   "Show your Ponto preferences",
	Long: `Shows the authenticated user's preferences: locale, theme, accent,
time_zone and export_locale.

time_zone is an IANA name (for example "America/Sao_Paulo") and is what a client
must use to decide which day an entry belongs to — the app groups a ledger by
started_at in that zone, not by the machine's clock.`,
	Example: "$ ponto preferences\n$ ponto preferences --jq '.data.time_zone'",
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := domainClient()
		if err != nil {
			return err
		}
		resp, err := c.Get("/preferences")
		if err != nil {
			return err
		}
		printDetail(resp.Data, "", []Breadcrumb{
			breadcrumb("status", "ponto auth status", "Check authentication"),
			breadcrumb("timer", "ponto timer status", "Show the running timer"),
		})
		return nil
	},
}

func init() {
	rootCmd.AddCommand(preferencesCmd)
}
