package cmd

import (
	"fmt"
	"os"

	"github.com/isai-arellano/kolyn-cli/cmd/ui"
	"github.com/spf13/cobra"
)

var (
	// Version will be set by goreleaser during build
	Version = "v0.2.1"
	Commit  = "none"
	Date    = "unknown"
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
	Version: Version,
}

var toolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "Herramientas de utilidad (Docker, etc)",
	Long:  `Colección de herramientas útiles para el desarrollo.`,
}

var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Gestiona servicios Docker",
	Long:  `Comandos para levantar y detener servicios Docker.`,
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(skillsCmd)
	rootCmd.AddCommand(toolsCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(syncCmd)

	toolsCmd.AddCommand(dockerCmd)
	toolsCmd.AddCommand(sshCmd)
	sshCmd.AddCommand(sshCreateCmd)

	dockerCmd.AddCommand(dockerUpCmd)
	dockerCmd.AddCommand(dockerDownCmd)
	dockerCmd.AddCommand(dockerListCmd)

	skillsCmd.AddCommand(skillsPathsCmd)
	skillsCmd.AddCommand(skillsListCmd)
}

func showWelcome() {
	ui.ShowBanner(Version)

	ui.Cyan.Println("📋 Comandos disponibles:")

	commands := []struct {
		name        string
		description string
	}{
		{"kolyn init", "Inicializa kolyn y agrega contexto al Agent.md"},
		{"kolyn sync", "Sincroniza skills remotos desde .kolyn.json"},
		{"kolyn skills", "Retorna JSON con skills disponibles para la IA"},
		{"kolyn skills list", "Lista skills y permite ver/editar contenido"},
		{"kolyn skills paths", "Retorna solo las rutas de skills"},
		{"kolyn update", "Actualiza kolyn a la última versión disponible"},
		{"kolyn tools ssh create", "Genera llaves SSH y configura acceso a servidores"},
		{"kolyn tools docker up", "Levanta servicios Docker desde templates"},
		{"kolyn tools docker list", "Lista servicios Docker y su estado"},
		{"kolyn tools docker down", "Detiene servicios Docker levantados"},
	}

	for _, cmd := range commands {
		ui.Blue.Printf("  🔹 %-25s", cmd.name)
		ui.WhiteText.Printf(" - %s\n", cmd.description)
	}

	fmt.Println()
	ui.YellowText.Println("💡 Tip: 'kolyn tools docker up' para levantar servicios Docker")
	fmt.Println()
}

// Execute ejecuta el comando raíz
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		ui.PrintError("%v", err)
		os.Exit(1)
	}
}
