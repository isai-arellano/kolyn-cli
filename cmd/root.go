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
	Long: `Kolyn es una herramienta que ayuda a agentes IA a:
- Obtener información y datos del proyecto
- Mantener contexto entre sesiones
- Acceder a skills y configuraciones

Usa 'kolyn <comando>' para interactuar.`,
	Run: func(cmd *cobra.Command, args []string) {
		showWelcome()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(skillsCmd)
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
	}

	for _, cmd := range commands {
		ui.Blue.Printf("  🔹 %-20s", cmd.name)
		ui.WhiteText.Printf(" - %s\n", cmd.description)
	}

	fmt.Println()
	ui.YellowText.Println("💡 Tip: Ejecuta 'kolyn init' para agregar contexto de kolyn al Agent.md")
	fmt.Println()
}

// Execute ejecuta el comando raíz
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		ui.PrintError("%v", err)
		os.Exit(1)
	}
}
