package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/isai-arellano/kolyn-cli/cmd/ui"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Muestra la versión actual de Kolyn",
	Run: func(cmd *cobra.Command, args []string) {
		ui.Cyan.Printf("Kolyn CLI version %s\n", Version)
		ui.Blue.Printf("Commit: %s\n", Commit)
		ui.Blue.Printf("Built at: %s\n", Date)
		fmt.Println()
	},
}
