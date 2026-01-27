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
	Short: "Inicializa kolyn y genera Agent.md",
	Long:  `Analiza el proyecto y genera un archivo Agent.md personalizado con las skills y reglas necesarias para la IA.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error obteniendo directorio actual: %w", err)
		}
		return RunInitProject(cmd.Context(), cwd, true)
	},
}

type ProjectFeatures struct {
	Type     string
	UI       bool
	Database bool
	Auth     bool
	API      bool
	DevOps   bool
}

// RunInitProject initializes a project at the given root directory.
// interactive: if true, asks the user for confirmation and features.
func RunInitProject(ctx context.Context, root string, interactive bool) error {
	ui.ShowSection("🚀 Inicializando Kolyn")

	// 1. Detección automática
	ui.PrintStep("Detectando tipo de proyecto...")
	pType := detectProjectType(root)
	ui.Cyan.Printf("   🔍 Tipo detectado: %s\n\n", strings.ToUpper(pType))

	agentPath := filepath.Join(root, "Agent.md")

	// 2. Advertencia si existe (solo en interactivo)
	if interactive {
		if _, err := os.Stat(agentPath); err == nil {
			ui.YellowText.Println("⚠️  Ya existe un archivo Agent.md en este proyecto.")
			ui.YellowText.Println("   Si continúas, se regenerará y perderás cambios manuales.")
			if !ui.AskYesNo("¿Deseas continuar?") {
				ui.PrintInfo("Operación cancelada.")
				return nil
			}
		}
	}

	// 3. Preguntas Interactivas o Defaults
	ui.PrintStep("Configurando features del proyecto:")

	features := ProjectFeatures{Type: pType}

	if interactive {
		features.UI = ui.AskYesNo("❓ ¿Tu proyecto usa componentes de UI? (shadcn, MUI, etc.)")
		features.Database = ui.AskYesNo("❓ ¿Tu proyecto tiene base de datos?")
		features.Auth = ui.AskYesNo("❓ ¿Tu proyecto tiene autenticación de usuarios?")
		features.API = ui.AskYesNo("❓ ¿Tu proyecto consume APIs externas?")
		features.DevOps = ui.AskYesNo("❓ ¿Tienes configurado CI/CD?")
	} else {
		// Defaults para modo no interactivo (ej. Scaffold)
		// Si es Next.js scaffold, asumimos UI (Shadcn) y API (Zod) por default
		if pType == "nextjs" {
			features.UI = true
			features.API = true
			// DB y Auth se dejan en false a menos que el scaffold diga lo contrario
		}
		ui.PrintInfo("Usando configuración por defecto para scaffold.")
	}

	// 4. Generar Agent.md
	if err := GenerateAgentMD(root, features); err != nil {
		return err
	}

	ui.Separator()
	ui.PrintSuccess("✅ Agent.md generado exitosamente.")
	ui.Gray.Println("Ahora puedes ejecutar 'kolyn check' para auditar el proyecto.")

	return nil
}

func detectProjectType(root string) string {
	// 1. Next.js
	if exists(filepath.Join(root, "next.config.ts")) ||
		exists(filepath.Join(root, "next.config.js")) ||
		exists(filepath.Join(root, "next.config.mjs")) {
		return "nextjs"
	}

	// 2. Go
	if exists(filepath.Join(root, "go.mod")) {
		return "go"
	}

	// 3. Python
	if exists(filepath.Join(root, "requirements.txt")) ||
		exists(filepath.Join(root, "pyproject.toml")) {
		return "python"
	}

	// 4. Node Generic
	if exists(filepath.Join(root, "package.json")) {
		return "node"
	}

	return "generic"
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// GenerateAgentMD generates the Agent.md file (exported for Scaffold use)
func GenerateAgentMD(root string, f ProjectFeatures) error {
	projectName := filepath.Base(root)

	// Construir lista de features activas
	var activeFeatures []string
	if f.UI {
		activeFeatures = append(activeFeatures, "ui")
	}
	if f.Database {
		activeFeatures = append(activeFeatures, "database")
	}
	if f.Auth {
		activeFeatures = append(activeFeatures, "auth")
	}
	if f.API {
		activeFeatures = append(activeFeatures, "api")
	}
	if f.DevOps {
		activeFeatures = append(activeFeatures, "devops")
	}

	// Construir lista de referencias sugeridas
	var refs []string

	// Base references based on type
	if f.Type == "nextjs" {
		refs = append(refs, "- [Next.js Framework](~/.kolyn/sources/github.com-isai-arellano-kolyn-skills/web/framework/nextjs.md)")
	}
	if f.Type == "go" {
		refs = append(refs, "- [Golang Core](~/.kolyn/sources/github.com-isai-arellano-kolyn-skills/backend/go/core.md)")
	}

	// Feature references
	if f.UI {
		refs = append(refs, "- [Shadcn UI](~/.kolyn/sources/github.com-isai-arellano-kolyn-skills/web/ui/shadcn.md)")
		refs = append(refs, "- [UI Stack](~/.kolyn/sources/github.com-isai-arellano-kolyn-skills/web/ui/stack.md)")
	}
	if f.Database {
		refs = append(refs, "- [Drizzle ORM](~/.kolyn/sources/github.com-isai-arellano-kolyn-skills/web/data/drizzle.md)")
		refs = append(refs, "- [PostgreSQL](~/.kolyn/sources/github.com-isai-arellano-kolyn-skills/web/data/postgres.md)")
	}
	if f.Auth {
		refs = append(refs, "- [Better Auth](~/.kolyn/sources/github.com-isai-arellano-kolyn-skills/web/auth/better-auth.md)")
	}
	if f.API || f.Type == "nextjs" { // Zod is almost always useful in web
		refs = append(refs, "- [Zod Validation](~/.kolyn/sources/github.com-isai-arellano-kolyn-skills/web/data/zod.md)")
	}
	if f.DevOps {
		refs = append(refs, "- [CI/CD](~/.kolyn/sources/github.com-isai-arellano-kolyn-skills/devops/ci-cd.md)")
	}

	content := fmt.Sprintf(`# Agent Context - %s

Kolyn Version: %s
Generated: %s
Project Type: %s

## Features
%s

---

## Project Context

### Stack & Architecture
This project uses the following stack conventions:
- **Type:** %s
%s

### Skills Reference
The following skills are active for this project. Use 'kolyn skills paths' to find more.

%s

### Rules
1. **Follow the Skills:** Read the reference files above before writing code.
2. **Directory Structure:** Respect the existing project structure.
3. **Consistency:** Use the same libraries and patterns defined in the stack.
`,
		projectName,
		Version,
		time.Now().Format("2006-01-02"),
		f.Type,
		formatList(activeFeatures),
		strings.ToUpper(f.Type),
		getStackSummary(f),
		strings.Join(refs, "\n"),
	)

	return os.WriteFile(filepath.Join(root, "Agent.md"), []byte(content), 0644)
}

func formatList(items []string) string {
	if len(items) == 0 {
		return "- core"
	}
	var sb strings.Builder
	for _, item := range items {
		sb.WriteString(fmt.Sprintf("- %s\n", item))
	}
	return strings.TrimRight(sb.String(), "\n")
}

func getStackSummary(f ProjectFeatures) string {
	var summary strings.Builder
	if f.UI {
		summary.WriteString("- **UI:** Shadcn/UI + Tailwind\n")
	}
	if f.Database {
		summary.WriteString("- **Database:** Drizzle ORM + PostgreSQL\n")
	}
	if f.Auth {
		summary.WriteString("- **Auth:** Better Auth\n")
	}
	return strings.TrimRight(summary.String(), "\n")
}
