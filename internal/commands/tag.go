package commands

import "github.com/spf13/cobra"

func tagFlags(cmd *cobra.Command) {
	cmd.Flags().String("name", "", "Tag name")
}

func init() {
	rootCmd.AddCommand(newCatalogCmd("tag", "tags", "tag", tagColumns, tagFlags, "name"))
}
