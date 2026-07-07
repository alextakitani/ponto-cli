package commands

import "github.com/spf13/cobra"

func clientFlags(cmd *cobra.Command) {
	cmd.Flags().String("name", "", "Client name")
	cmd.Flags().String("currency", "", "Currency code")
	cmd.Flags().Int("rate-cents", 0, "Hourly rate in cents")
	cmd.Flags().String("note", "", "Client note")
}

func init() {
	rootCmd.AddCommand(newCatalogCmd("client", "clients", "client", clientColumns, clientFlags, "name", "currency"))
}
