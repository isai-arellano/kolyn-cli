package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/isai-arellano/kolyn-cli/cmd/ui"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Inicializa kolyn y agrega contexto al Agent.md",
	Long:  `Agrega información de kolyn al Agent.md para que la IA tenga contexto de cómo usar kolyn CLI.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInitCommand(cmd.Context())
	},
}

const kolynContextTemplate = `
═══════════════════════════════════════════════════════════════════════
KOLYN CONTEXT & TOOLS
═══════════════════════════════════════════════════════════════════════

⚠️ TOKEN ECONOMY NOTICE:
No leas todas las skills de golpe. Usa 'kolyn skills paths' para ver el índice y lee SOLO el archivo específico que necesites para la tarea actual.

🛠 COMMANDS:
• kolyn skills paths        → Muestra rutas de skills (Índice Maestro)
• kolyn check               → Audita que el proyecto cumpla con las skills (Deps, Files)
• kolyn tools docker list   → Ver servicios corriendo
• kolyn tools docker up     → Levantar infraestructura (DBs, n8n, etc)

📌 SKILL MAP (Si vas a tocar X, lee Y):
• 🎨 UI / Components      → Lee skills/web/ui/ (shadcn.md, stack.md)
   ↳ Requisito: Framer Motion, React Icons, Sonner, Tailwind+CVA.
• 🔐 Auth / Sessions      → Lee skills/web/auth/ (better-auth.md)
   ↳ Requisito: Better Auth, Plugins, Secure Cookies.
• 💾 Data / DB / Schema   → Lee skills/web/data/ (drizzle.md, postgres.md, zod.md)
   ↳ Requisito: Drizzle ORM, Postgres 3NF, Zod Validation.
• ⚡ Framework / Logic    → Lee skills/web/framework/ (nextjs.md)
   ↳ Requisito: Next.js 16, Server Actions, 'use client' en hojas.
• 🐹 Backend / Golang     → Lee skills/golang/core.md
   ↳ Requisito: Go 1.22+, errgroup, estructura cmd/internal.

═══════════════════════════════════════════════════════════════════════
`

func runInitCommand(ctx context.Context) error {
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
		if err := updateAgentMD(agentPath); err != nil {
			ui.PrintWarning("No se pudo actualizar Agent.md: %v", err)
		} else {
			ui.PrintSuccess("Agent.md actualizado")
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

func updateAgentMD(agentPath string) error {
	content, err := os.ReadFile(agentPath)
	if err != nil {
		return fmt.Errorf("error leyendo Agent.md: %w", err)
	}

	contentStr := string(content)

	if strings.Contains(contentStr, "KOLYN") {
		ui.PrintStep("Actualizando contexto de kolyn en Agent.md...")
		newContent := removeKolynBlock(contentStr)
		newContent = strings.TrimRight(newContent, "\n") + "\n" + kolynContextTemplate
		if err := os.WriteFile(agentPath, []byte(newContent), 0644); err != nil {
			return fmt.Errorf("error escribiendo Agent.md: %w", err)
		}
		return nil
	}

	ui.PrintStep("Agregando contexto de kolyn al Agent.md...")
	newContent := contentStr + "\n" + kolynContextTemplate
	if err := os.WriteFile(agentPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("error escribiendo Agent.md: %w", err)
	}

	return nil
}

func removeKolynBlock(content string) string {
	lines := strings.Split(content, "\n")
	var result []string
	skipMode := false
	currentIdx := -1

	for _, line := range lines {
		currentIdx++
		trimmed := strings.TrimSpace(line)

		// Detect start of block (heuristic)
		if trimmed == "KOLYN" || (strings.Contains(line, "═") && strings.Contains(line, "KOLYN")) {
			skipMode = true
			continue
		}

		if skipMode {
			// Check for end of block: a line starting with ═ that is not the start
			if strings.HasPrefix(trimmed, "═") && trimmed != "" {
				// Look ahead to see if this is truly the end (followed by non-empty content not part of block)
				// Or just assume it closes the block.
				// The original logic was complex. Let's simplify:
				// The block ends with a separator line.
				skipMode = false
				continue
			}
			continue
		}

		result = append(result, line)
	}

	return strings.TrimRight(strings.Join(result, "\n"), "\n")
}

func createAgentWithKolyn(projectPath string) error {
	projectName := filepath.Base(projectPath)

	agentContent := fmt.Sprintf(`# Agent Context - %s

Kolyn Version: %s
Creado: %s

%s`, projectName, Version, time.Now().Format("2006-01-02"), kolynContextTemplate)

	agentPath := filepath.Join(projectPath, "Agent.md")

	if err := os.WriteFile(agentPath, []byte(agentContent), 0644); err != nil {
		return fmt.Errorf("error creando Agent.md: %w", err)
	}

	return nil
}
