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

	// 3. Verificar si existen skills (si no, advertir)
	skillsDir, err := getFirstSkillsSourceDir()
	if err != nil || skillsDir == "" {
		ui.YellowText.Println("\n⚠️  No se detectaron skills sincronizadas.")
		ui.Gray.Println("   El Agent.md generado será básico y no tendrá referencias a skills.")
		ui.Gray.Println("   Recomendación: Ejecuta 'kolyn sync' después de esto.")
		if interactive && !ui.AskYesNo("¿Deseas continuar de todos modos?") {
			return nil
		}
	}

	// 4. Preguntas Interactivas o Defaults
	ui.PrintStep("Configurando features del proyecto:")

	features := ProjectFeatures{Type: pType}

	if interactive {
		features.UI = ui.AskYesNo("❓ ¿Tu proyecto usa componentes de UI?")
		features.Database = ui.AskYesNo("❓ ¿Tu proyecto tiene base de datos?")
		features.Auth = ui.AskYesNo("❓ ¿Tu proyecto tiene autenticación de usuarios?")
		features.API = ui.AskYesNo("❓ ¿Tu proyecto consume APIs externas?")
		features.DevOps = ui.AskYesNo("❓ ¿Tienes configurado CI/CD?")
	} else {
		// Defaults para modo no interactivo (ej. Scaffold)
		if pType == "nextjs" {
			features.UI = true
			features.API = true
		}
		ui.PrintInfo("Usando configuración por defecto para scaffold.")
	}

	// 5. Generar Agent.md
	if err := GenerateAgentMD(root, features, skillsDir); err != nil {
		return err
	}

	ui.Separator()
	ui.PrintSuccess("✅ Agent.md generado exitosamente.")
	ui.Gray.Println("Ahora puedes ejecutar 'kolyn check' para auditar el proyecto.")

	return nil
}

func detectProjectType(root string) string {
	if exists(filepath.Join(root, "next.config.ts")) ||
		exists(filepath.Join(root, "next.config.js")) ||
		exists(filepath.Join(root, "next.config.mjs")) {
		return "nextjs"
	}
	if exists(filepath.Join(root, "go.mod")) {
		return "go"
	}
	if exists(filepath.Join(root, "requirements.txt")) ||
		exists(filepath.Join(root, "pyproject.toml")) {
		return "python"
	}
	if exists(filepath.Join(root, "package.json")) {
		return "node"
	}
	return "generic"
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// GenerateAgentMD generates the Agent.md file
func GenerateAgentMD(root string, f ProjectFeatures, skillsDirName string) error {
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

	// Contenido base del archivo
	var content strings.Builder

	// Header
	fmt.Fprintf(&content, `# Agent Context - %s

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
- **Capabilities:** %s
`,
		projectName,
		Version,
		time.Now().Format("2006-01-02"),
		f.Type,
		formatList(activeFeatures),
		strings.ToUpper(f.Type),
		formatInlineList(activeFeatures),
	)

	// Solo agregar sección de Skills si tenemos skills sincronizadas
	if skillsDirName != "" {
		basePath := fmt.Sprintf("~/.kolyn/sources/%s", skillsDirName)
		var refs []string

		// Base references based on type
		if f.Type == "nextjs" {
			refs = append(refs, fmt.Sprintf("- [Next.js Framework](%s/web/framework/nextjs.md)", basePath))
		}
		if f.Type == "go" {
			refs = append(refs, fmt.Sprintf("- [Golang Core](%s/backend/go/core.md)", basePath))
		}

		// Feature references
		if f.UI {
			refs = append(refs, fmt.Sprintf("- [UI Components](%s/web/ui/shadcn.md)", basePath))
			refs = append(refs, fmt.Sprintf("- [UI Stack](%s/web/ui/stack.md)", basePath))
		}
		if f.Database {
			refs = append(refs, fmt.Sprintf("- [Database/ORM](%s/web/data/drizzle.md)", basePath))
			refs = append(refs, fmt.Sprintf("- [Database Design](%s/web/data/postgres.md)", basePath))
		}
		if f.Auth {
			refs = append(refs, fmt.Sprintf("- [Authentication](%s/web/auth/better-auth.md)", basePath))
		}
		if f.API || f.Type == "nextjs" {
			refs = append(refs, fmt.Sprintf("- [Data Validation](%s/web/data/zod.md)", basePath))
		}
		if f.DevOps {
			refs = append(refs, fmt.Sprintf("- [CI/CD](%s/devops/ci-cd.md)", basePath))
		}

		if len(refs) > 0 {
			fmt.Fprintf(&content, `
### Skills Reference
The following skills are active for this project. Use 'kolyn skills paths' to find more.

%s
`, strings.Join(refs, "\n"))
		}
	} else {
		// Mensaje alternativo si no hay skills
		fmt.Fprintf(&content, `
### Skills Reference
⚠️ No skills detected. Run 'kolyn sync' to download your team's skills and then regenerate this file with 'kolyn init'.
`)
	}

	// Footer con reglas generales
	fmt.Fprintf(&content, `
### Rules
1. **Follow the Skills:** Read the reference files above before writing code.
2. **Directory Structure:** Respect the existing project structure.
3. **Consistency:** Use the same libraries and patterns defined in the stack.
`)

	return os.WriteFile(filepath.Join(root, "Agent.md"), []byte(content.String()), 0644)
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

func formatInlineList(items []string) string {
	if len(items) == 0 {
		return "Core only"
	}
	return strings.Join(items, ", ")
}

// getFirstSkillsSourceDir intenta encontrar el primer directorio de skills disponible
func getFirstSkillsSourceDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	sourcesDir := filepath.Join(home, ".kolyn", "sources")

	entries, err := os.ReadDir(sourcesDir)
	if err != nil {
		return "", err // Probablemente no existe
	}

	for _, entry := range entries {
		if entry.IsDir() {
			return entry.Name(), nil
		}
	}
	return "", nil
}
