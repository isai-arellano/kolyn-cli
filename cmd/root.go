package cmd

import (
	"fmt"
	"os"

	"github.com/isai-arellano/kolyn-cli/cmd/ui"
	"github.com/spf13/cobra"
)

var (
	// Version will be set by goreleaser during build
	Version = "v0.2.15"
	Commit  = "none"
	Date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "kolyn",
	Short: "Orquestador de desarrollo para equipos con IA",
	Long: `Kolyn es una herramienta CLI diseñada para estandarizar flujos de trabajo
y proveer contexto a agentes de IA.

Usa 'kolyn <comando>' para interactuar.`,
	Run: func(cmd *cobra.Command, args []string) {
		showWelcome()
	},
	Version: Version,
}

var toolsCmd = &cobra.Command{
	Use:   "tools",
	Short: "Herramientas de utilidad (Docker, SSH, etc)",
	Long:  `Colección de herramientas útiles para el desarrollo.`,
}

var dockerCmd = &cobra.Command{
	Use:   "docker",
	Short: "Gestiona servicios Docker",
	Long:  `Comandos para levantar y detener servicios Docker.`,
}

func init() {
	// Comandos Principales (Root)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(scaffoldCmd)
	rootCmd.AddCommand(configCmd)

	// Comandos de Servicios (Promovidos a Root)
	rootCmd.AddCommand(dockerUpCmd)   // kolyn up
	rootCmd.AddCommand(dockerDownCmd) // kolyn down
	rootCmd.AddCommand(statusCmd)     // kolyn status

	// Comandos Utilitarios
	rootCmd.AddCommand(skillsCmd)
	rootCmd.AddCommand(toolsCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(uninstallCmd)

	// Estructura Legacy/Organizada
	toolsCmd.AddCommand(dockerCmd)
	toolsCmd.AddCommand(sshCmd)
	sshCmd.AddCommand(sshCreateCmd)

	dockerCmd.AddCommand(dockerUpCmd)
	dockerCmd.AddCommand(dockerDownCmd)
	dockerCmd.AddCommand(statusCmd)

	skillsCmd.AddCommand(skillsPathsCmd)
	skillsCmd.AddCommand(skillsListCmd)
}

func showWelcome() {
	ui.ShowBanner(Version)

	ui.Cyan.Println("📋 Proyecto:")

	commands := []struct {
		name        string
		description string
	}{
		{"kolyn init", "Inicializa proyecto con contexto para IA"},
		{"kolyn sync", "Sincroniza skills del equipo"},
		{"kolyn check", "Audita cumplimiento de estándares"},
		{"kolyn scaffold", "Genera estructura base de proyectos"},
		{"kolyn config", "Configuración global (idioma, repo)"},
	}

	for _, cmd := range commands {
		ui.Blue.Printf("  🔹 %-25s", cmd.name)
		ui.WhiteText.Printf(" - %s\n", cmd.description)
	}

	fmt.Println()
	ui.Cyan.Println("🐳 Servicios:")

	services := []struct {
		name        string
		description string
	}{
		{"kolyn up", "Levanta servicios Docker"},
		{"kolyn down", "Detiene servicios"},
		{"kolyn status", "Ver estado de servicios"},
	}

	for _, cmd := range services {
		ui.Blue.Printf("  🔹 %-25s", cmd.name)
		ui.WhiteText.Printf(" - %s\n", cmd.description)
	}

	fmt.Println()
	ui.YellowText.Println("💡 Tip: Usa 'kolyn <command> --help' para más detalles.")
	fmt.Println()
}

// Execute ejecuta el comando raíz
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		ui.PrintError("%v", err)
		os.Exit(1)
	}
}
