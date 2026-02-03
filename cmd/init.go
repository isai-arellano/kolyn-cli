package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/isai-arellano/kolyn-cli/cmd/ui"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Inicializa kolyn y genera Agent.md",
	Long:  `Analiza el proyecto y genera un archivo Agent.md personalizado con las skills seleccionadas.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error obteniendo directorio actual: %w", err)
		}
		return RunInitProject(cmd.Context(), cwd, true)
	},
}

// RunInitProject initializes a project at the given root directory.
func RunInitProject(ctx context.Context, root string, interactive bool) error {
	ui.ShowSection("🚀 Inicializando Kolyn")

	// 1. Detección automática (solo metadata)
	ui.PrintStep("Detectando tipo de proyecto...")
	pType := detectProjectType(root)
	ui.Cyan.Printf("   🔍 Tipo base: %s\n\n", strings.ToUpper(pType))

	agentPath := filepath.Join(root, "Agent.md")
	var existingSkills map[string]bool
	var err error

	// 2. Leer skills existentes si ya hay Agent.md
	if exists(agentPath) {
		ui.PrintInfo("Agent.md existente detectado. Leyendo configuración actual...")
		existingSkills, err = readExistingSkillsFromAgent(agentPath)
		if err != nil {
			ui.PrintWarning(fmt.Sprintf("No se pudieron leer las skills actuales: %v", err))
		}
	} else {
		existingSkills = make(map[string]bool)
	}

	// 3. Escanear skills disponibles
	allSkills, err := scanSkills(ctx)
	if err != nil {
		ui.YellowText.Println("\n⚠️  No se pudieron escanear las skills.")
		ui.Gray.Println("   Asegúrate de ejecutar 'kolyn sync' primero.")
		if interactive && !ui.AskYesNo("¿Deseas continuar sin skills?") {
			return nil
		}
	}

	// 4. Selección Interactiva (Huh)
	var selectedSkills []SkillInfo

	if interactive && len(allSkills) > 0 {
		// Agrupar skills para presentación visual (aunque huh flat list también se ve bien)
		// Vamos a usar el formato "Category › Name" para las opciones de huh

		// Ordenar todo por categoría y nombre
		sort.Slice(allSkills, func(i, j int) bool {
			if allSkills[i].Category == allSkills[j].Category {
				return allSkills[i].Name < allSkills[j].Name
			}
			return allSkills[i].Category < allSkills[j].Category
		})

		var uiOptions []ui.SkillOption
		skillMap := make(map[string]SkillInfo) // Para recuperar el objeto SkillInfo después

		for _, s := range allSkills {
			// Construir label bonito
			label := fmt.Sprintf("%s › %s", s.Category, s.Name)
			if s.Category == "root" || s.Category == "." {
				label = s.Name
			}

			// Determinar si estaba seleccionado
			isSelected := isSkillSelected(s.Path, existingSkills)

			uiOptions = append(uiOptions, ui.SkillOption{
				Label:       label,
				Value:       s.Path,
				Description: s.Description,
				Selected:    isSelected,
			})
			skillMap[s.Path] = s
		}

		// Llamar al nuevo selector
		selectedPaths, err := ui.SelectSkills("Selecciona las skills para este proyecto:", uiOptions)
		if err != nil {
			return nil // Cancelado
		}

		// Reconstruir lista de selectedSkills
		for _, path := range selectedPaths {
			if skill, ok := skillMap[path]; ok {
				selectedSkills = append(selectedSkills, skill)
			}
		}

	} else if len(allSkills) > 0 {
		ui.PrintInfo("Modo no interactivo: No se seleccionaron skills adicionales.")
	}

	// 5. Generar Agent.md
	if err := GenerateAgentMD(root, pType, selectedSkills); err != nil {
		return err
	}

	ui.Separator()
	ui.PrintSuccess("✅ Agent.md generado exitosamente.")
	ui.Gray.Printf("   Skills activas: %d\n", len(selectedSkills))
	ui.Gray.Println("Ahora puedes ejecutar 'kolyn check' para auditar el proyecto.")

	return nil
}

type CategoryGroup struct {
	Name   string
	Skills []SkillInfo
}

// isSkillSelected verifica si un path de skill está en el map de existentes.
func isSkillSelected(skillPath string, existing map[string]bool) bool {
	// Intentar match exacto
	if existing[skillPath] {
		return true
	}
	// Intentar match por nombre de archivo
	base := filepath.Base(skillPath)
	for k := range existing {
		if strings.Contains(k, base) {
			return true
		}
	}
	return false
}

// readExistingSkillsFromAgent parsea el Agent.md y busca links en la sección de Skills
func readExistingSkillsFromAgent(path string) (map[string]bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	skills := make(map[string]bool)
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	inSkillsSection := false

	// Regex para markdown links: [Title](path)
	linkRegex := regexp.MustCompile(`\[.*?\]\((.*?)\)`)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "### Skills Reference") {
			inSkillsSection = true
			continue
		}
		if strings.HasPrefix(line, "### ") && inSkillsSection {
			inSkillsSection = false // Fin de sección
			break
		}

		if inSkillsSection {
			matches := linkRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				linkPath := matches[1]
				skills[linkPath] = true
			}
		}
	}
	return skills, nil
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

// GenerateAgentMD generates the Agent.md file with selected skills
func GenerateAgentMD(root string, pType string, skills []SkillInfo) error {
	projectName := filepath.Base(root)

	var content strings.Builder

	// Header
	fmt.Fprintf(&content, `# Agent Context - %s

Kolyn Version: %s
Generated: %s
Project Type: %s

---

## Project Context

### Stack & Architecture
This project is defined by the following selected skills.
Type: %s
`,
		projectName,
		Version, // Asumiendo que Version es global en cmd package, si no, habría que importarlo o definirlo
		time.Now().Format("2006-01-02"),
		pType,
		strings.ToUpper(pType),
	)

	// Skills Section
	if len(skills) > 0 {
		fmt.Fprintf(&content, "\n### Skills Reference\nThe following skills are active for this project.\n\n")

		home, _ := os.UserHomeDir()

		for _, s := range skills {
			displayPath := s.Path
			if strings.HasPrefix(s.Path, home) {
				displayPath = strings.Replace(s.Path, home, "~", 1)
			}

			// Formato: - [Name](path)
			fmt.Fprintf(&content, "- [%s (%s)](%s)\n", s.Name, s.Category, displayPath)
		}
	} else {
		fmt.Fprintf(&content, `
### Skills Reference
⚠️ No skills selected. Run 'kolyn init' again to add skills.
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

// Helper to get first skills dir (retained for backward compat if needed, mainly used by scanSkills now)
func getFirstSkillsSourceDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	sourcesDir := filepath.Join(home, ".kolyn", "sources")

	entries, err := os.ReadDir(sourcesDir)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			return entry.Name(), nil
		}
	}
	return "", nil
}
