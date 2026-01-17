package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourusername/kolyn/cmd/ui"
)

var rootCmd = &cobra.Command{
	Use:   "kolyn",
	Short: "Herramienta CLI para ayudar a la IA con contexto y datos",
	Long: `Kolyn es una herramienta CLI que:
- Agrega contexto del proyecto para agentes IA
- Proporciona acceso a skills y templates
- Permite levantar servicios Docker rápidamente

Usa 'kolyn <comando>' para interactuar.`,
	Run: func(cmd *cobra.Command, args []string) {
		showWelcome()
	},
}

var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Gestiona servicios Docker",
	Long:  `Comandos para levantar y detener servicios Docker.`,
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(skillsCmd)
	rootCmd.AddCommand(dockerCmd)
	dockerCmd.AddCommand(dockerUpCmd)
	dockerCmd.AddCommand(dockerDownCmd)
	dockerCmd.AddCommand(dockerListCmd)
	skillsCmd.AddCommand(skillsPathsCmd)
	skillsCmd.AddCommand(skillsListCmd)
}

func showWelcome() {
	ui.ShowBanner()

	ui.Cyan.Println("📋 Comandos disponibles:")

	commands := []struct {
		name        string
		description string
	}{
		{"kolyn init", "Inicializa kolyn y agrega contexto al Agent.md"},
		{"kolyn skills", "Retorna JSON con skills disponibles para la IA"},
		{"kolyn skills list", "Lista skills y permite ver/editar contenido"},
		{"kolyn skills paths", "Retorna solo las rutas de skills"},
		{"kolyn docker up", "Levanta servicios Docker desde templates"},
		{"kolyn docker list", "Lista servicios Docker y su estado"},
		{"kolyn docker down", "Detiene servicios Docker levantados"},
	}

	for _, cmd := range commands {
		ui.Blue.Printf("  🔹 %-20s", cmd.name)
		ui.WhiteText.Printf(" - %s\n", cmd.description)
	}

	fmt.Println()
	ui.YellowText.Println("💡 Tip: 'kolyn docker up' para levantar servicios Docker")
	fmt.Println()
}

// Execute ejecuta el comando raíz
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		ui.PrintError("%v", err)
		os.Exit(1)
	}
}
