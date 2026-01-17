package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourusername/kolyn/cmd/ui"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Inicializa kolyn y agrega contexto al Agent.md",
	Long:  `Agrega información de kolyn al Agent.md para que la IA tenga contexto de cómo usar kolyn CLI.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return initializeKolyn()
	},
}

// KolynContext es el contenido conciso para Agent.md
const kolynContextTemplate = `

═══════════════════════════════════════════════════════════════════════
KOLYN
═══════════════════════════════════════════════════════════════════════

kolyn init           → Inicializa kolyn en el proyecto
kolyn skills         → JSON con skills disponibles para la IA
kolyn skills list    → Lista skills y permite ver/editar contenido
kolyn skills paths   → Retorna solo las rutas de skills
kolyn docker up      → Levanta servicios Docker (n8n, postgres, etc.)
kolyn docker list    → Lista servicios Docker y su estado
kolyn docker down    → Detiene servicios Docker levantados

═══════════════════════════════════════════════════════════════════════
`

// initializeKolyn inicializa kolyn en el proyecto actual
func initializeKolyn() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error obteniendo directorio actual: %w", err)
	}

	ui.ShowSection("🚀 Inicializando Kolyn")

	agentPath := filepath.Join(cwd, "Agent.md")

	agentInfo, err := os.Stat(agentPath)

	if err == nil && agentInfo.IsDir() {
		ui.PrintWarning("Agent.md es un directorio")
		return nil
	} else if err == nil && !agentInfo.IsDir() {
		ui.PrintStep("Agregando contexto de kolyn al Agent.md...")
		if err := addKolynContextToAgent(agentPath); err != nil {
			ui.PrintWarning("No se pudo actualizar Agent.md: %v", err)
		} else {
			ui.PrintSuccess("Contexto de kolyn agregado al Agent.md")
		}
	} else {
		ui.PrintStep("Creando Agent.md con contexto de kolyn...")
		if err := createAgentWithKolyn(cwd); err != nil {
			return fmt.Errorf("error creando Agent.md: %w", err)
		}
		ui.PrintSuccess("Agent.md creado")
	}

	ui.Separator()
	ui.PrintSuccess("Kolyn inicializado!")

	return nil
}

// addKolynContextToAgent agrega contexto de kolyn al Agent.md existente
func addKolynContextToAgent(agentPath string) error {
	content, err := os.ReadFile(agentPath)
	if err != nil {
		return fmt.Errorf("error leyendo Agent.md: %w", err)
	}

	if strings.Contains(string(content), "KOLYN") {
		ui.PrintInfo("Agent.md ya tiene contexto de kolyn")
		return nil
	}

	newContent := string(content) + kolynContextTemplate

	if err := os.WriteFile(agentPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("error escribiendo Agent.md: %w", err)
	}

	return nil
}

// createAgentWithKolyn crea un nuevo Agent.md con contexto de kolyn
func createAgentWithKolyn(projectPath string) error {
	projectName := filepath.Base(projectPath)

	agentContent := fmt.Sprintf(`# Agent Context - %s

Kolyn Version: %s
Creado: %s

%s`, projectName, "v0.3.0", time.Now().Format("2006-01-02"), kolynContextTemplate)

	agentPath := filepath.Join(projectPath, "Agent.md")

	if err := os.WriteFile(agentPath, []byte(agentContent), 0644); err != nil {
		return fmt.Errorf("error creando Agent.md: %w", err)
	}

	return nil
}
