package cmd

import (
	"bufio"
	"bytes"
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
	"gopkg.in/yaml.v3"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Inicializa kolyn y genera Agent.md",
	Long:  `Analiza el proyecto, copia las skills seleccionadas a .kolyn/skills/ y genera un archivo Agent.md con reglas inyectadas.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error obteniendo directorio actual: %w", err)
		}
		return RunInitProject(cmd.Context(), cwd, true)
	},
}

// Internal struct to hold skill data during init process
type SelectedSkillData struct {
	OriginalPath string
	LocalPath    string // Path relative to project root (e.g. .kolyn/skills/foo.md)
	Name         string
	Category     string
	Rules        []string
}

// SkillFrontmatter structure reused for extraction
type SkillFrontmatterInit struct {
	Name       string   `yaml:"name"`
	AgentRules []string `yaml:"agent_rules"`
}

// RunInitProject initializes a project at the given root directory.
func RunInitProject(ctx context.Context, root string, interactive bool) error {
	ui.ShowSection("🚀 Inicializando Kolyn")

	// 1. Detección automática
	ui.PrintStep("Detectando tipo de proyecto...")
	pType := detectProjectType(root)
	ui.Cyan.Printf("   🔍 Tipo base: %s\n\n", strings.ToUpper(pType))

	agentPath := filepath.Join(root, "Agent.md")
	var existingSkills map[string]bool
	var err error

	// 2. Leer skills existentes
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
		if interactive && !ui.AskYesNo("¿Deseas continuar sin skills?") {
			return nil
		}
	}

	// 4. Selección Interactiva
	var selectedSkillsRaw []SkillInfo

	if interactive && len(allSkills) > 0 {
		sort.Slice(allSkills, func(i, j int) bool {
			if allSkills[i].Category == allSkills[j].Category {
				return allSkills[i].Name < allSkills[j].Name
			}
			return allSkills[i].Category < allSkills[j].Category
		})

		var uiOptions []ui.SkillOption
		skillMap := make(map[string]SkillInfo)

		for _, s := range allSkills {
			label := fmt.Sprintf("%s › %s", s.Category, s.Name)
			if s.Category == "root" || s.Category == "." {
				label = s.Name
			}

			isSelected := isSkillSelected(s.Path, existingSkills)

			uiOptions = append(uiOptions, ui.SkillOption{
				Label:       label,
				Value:       s.Path,
				Description: s.Description,
				Selected:    isSelected,
			})
			skillMap[s.Path] = s
		}

		selectedPaths, err := ui.SelectSkills("Selecciona las skills para este proyecto:", uiOptions)
		if err != nil {
			return nil // Cancelado
		}

		for _, path := range selectedPaths {
			if skill, ok := skillMap[path]; ok {
				selectedSkillsRaw = append(selectedSkillsRaw, skill)
			}
		}

	} else if len(allSkills) > 0 {
		ui.PrintInfo("Modo no interactivo: No se seleccionaron skills adicionales.")
	}

	// 5. Copiar Skills y Extraer Reglas (Vendorización)
	if len(selectedSkillsRaw) > 0 {
		ui.PrintStep("Vendorizando skills y extrayendo reglas...")

		skillsDestDir := filepath.Join(root, ".kolyn", "skills")
		if err := os.MkdirAll(skillsDestDir, 0755); err != nil {
			return fmt.Errorf("error creando directorio de skills: %w", err)
		}

		for _, skill := range selectedSkillsRaw {
			localPath, _, err := copySkillToProject(skill.Path, root, skillsDestDir)
			if err != nil {
				ui.PrintError("Fallo al copiar skill %s: %v", skill.Name, err)
				continue
			}
			ui.Gray.Printf("   ✅ %s -> %s\n", skill.Name, localPath)
		}
	}

	// 5.5 Recargar TODAS las skills locales (nuevas + antiguas) para generar el Agent.md completo
	allLocalSkills, err := loadAllLocalSkills(root)
	if err != nil {
		ui.PrintWarning(fmt.Sprintf("Advertencia: No se pudieron recargar las skills locales: %v", err))
	}

	// 6. Generar o Actualizar Agent.md
	if err := GenerateAgentMD(root, pType, allLocalSkills); err != nil {
		return err
	}

	ui.Separator()
	if len(selectedSkillsRaw) > 0 {
		ui.PrintSuccess("✅ Agent.md actualizado con nuevas skills.")
	} else {
		ui.PrintSuccess("✅ Agent.md regenerado/verificado.")
	}
	ui.Gray.Printf("   Total skills activas: %d\n", len(allLocalSkills))
	ui.Gray.Println("Ahora el proyecto es autónomo. Las skills viven en .kolyn/skills/")

	return nil
}

// loadAllLocalSkills lee todas las skills en .kolyn/skills para reconstruir el estado completo
func loadAllLocalSkills(root string) ([]SelectedSkillData, error) {
	skillsDir := filepath.Join(root, ".kolyn", "skills")
	var results []SelectedSkillData

	entries, err := os.ReadDir(skillsDir)
	if os.IsNotExist(err) {
		return results, nil
	}
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		fullPath := filepath.Join(skillsDir, entry.Name())
		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		// Parse Frontmatter
		var name string = strings.TrimSuffix(entry.Name(), ".md")
		var rules []string
		var category string = "Installed" // Default category since folder structure is flattened

		if bytes.HasPrefix(content, []byte("---")) {
			parts := bytes.SplitN(content, []byte("---"), 3)
			if len(parts) >= 3 {
				var fm SkillFrontmatterInit
				if err := yaml.Unmarshal(parts[1], &fm); err == nil {
					if fm.Name != "" {
						name = fm.Name
					}
					rules = fm.AgentRules
				}
			}
		}

		relPath, _ := filepath.Rel(root, fullPath)
		relPath = "./" + filepath.ToSlash(relPath)

		results = append(results, SelectedSkillData{
			OriginalPath: fullPath,
			LocalPath:    relPath,
			Name:         name,
			Category:     category,
			Rules:        rules,
		})
	}
	return results, nil
}

// copySkillToProject copia el archivo, extrae reglas y devuelve el path relativo
func copySkillToProject(srcPath, root, destDir string) (string, []string, error) {
	content, err := os.ReadFile(srcPath)
	if err != nil {
		return "", nil, err
	}

	var rules []string
	if bytes.HasPrefix(content, []byte("---")) {
		parts := bytes.SplitN(content, []byte("---"), 3)
		if len(parts) >= 3 {
			var fm SkillFrontmatterInit
			if err := yaml.Unmarshal(parts[1], &fm); err == nil {
				rules = fm.AgentRules
			}
		}
	}

	baseName := filepath.Base(srcPath)
	destPath := filepath.Join(destDir, baseName)

	if err := os.WriteFile(destPath, content, 0644); err != nil {
		return "", nil, err
	}

	relPath, err := filepath.Rel(root, destPath)
	if err != nil {
		relPath = destPath
	}

	relPath = filepath.ToSlash(relPath)
	if !strings.HasPrefix(relPath, ".") {
		relPath = "./" + relPath
	}

	return relPath, rules, nil
}

func isSkillSelected(skillPath string, existing map[string]bool) bool {
	if existing[skillPath] {
		return true
	}
	base := filepath.Base(skillPath)
	for k := range existing {
		if filepath.Base(k) == base {
			return true
		}
	}
	return false
}

func readExistingSkillsFromAgent(path string) (map[string]bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	skills := make(map[string]bool)
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	inSkillsSection := false

	linkRegex := regexp.MustCompile(`\[.*?\]\((.*?)\)`)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "### Skills Reference") {
			inSkillsSection = true
			continue
		}
		if strings.HasPrefix(line, "### ") && inSkillsSection {
			inSkillsSection = false
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

func GenerateAgentMD(root string, pType string, skills []SelectedSkillData) error {
	agentPath := filepath.Join(root, "Agent.md")

	// Generar el contenido de las secciones dinámicas
	var skillsBlock strings.Builder
	var rulesBlock strings.Builder

	// Ordenar skills por nombre para consistencia
	sort.Slice(skills, func(i, j int) bool {
		return skills[i].Name < skills[j].Name
	})

	if len(skills) > 0 {
		skillsBlock.WriteString("\nThe following skills are active for this project.\n\n")
		for _, s := range skills {
			// Intentar mantener la categoría si es posible, o usar "Skill"
			cat := s.Category
			if cat == "" {
				cat = "Skill"
			}
			skillsBlock.WriteString(fmt.Sprintf("- [%s (%s)](%s)\n", s.Name, cat, s.LocalPath))
		}
	} else {
		skillsBlock.WriteString("\n⚠️ No skills selected. Run 'kolyn init' again to add skills.\n")
	}

	ruleCounter := 1
	for _, s := range skills {
		if len(s.Rules) > 0 {
			rulesBlock.WriteString(fmt.Sprintf("\n#### From %s:\n", s.Name))
			for _, r := range s.Rules {
				rulesBlock.WriteString(fmt.Sprintf("%d. %s\n", ruleCounter, r))
				ruleCounter++
			}
		}
	}
	rulesBlock.WriteString("\n#### General:\n")
	rulesBlock.WriteString(fmt.Sprintf("%d. **Follow the Skills:** Read the reference files above before writing code.\n", ruleCounter))
	ruleCounter++
	rulesBlock.WriteString(fmt.Sprintf("%d. **Directory Structure:** Respect the existing project structure.\n", ruleCounter))
	ruleCounter++
	rulesBlock.WriteString(fmt.Sprintf("%d. **Consistency:** Use the same libraries and patterns defined in the stack.\n", ruleCounter))

	// Leer archivo existente para intentar "hidratar"
	existingContent, err := os.ReadFile(agentPath)
	if err == nil && len(existingContent) > 0 {
		contentStr := string(existingContent)

		// Regex para encontrar los bloques y reemplazarlos
		// Buscamos: (Todo antes de Skills) (Header Skills + Contenido) (Header Rules) (Contenido Rules) (Resto)
		// Nota: Asumimos que "### Rules" viene después de "### Skills Reference"

		// 1. Reemplazar sección de Skills
		skillsRegex := regexp.MustCompile(`(?s)(### Skills Reference).*?(### Rules)`)
		loc := skillsRegex.FindStringIndex(contentStr)

		if loc != nil {
			// Encontramos el bloque estándar. Procedemos a reemplazar secciones.
			newContent := contentStr

			// Reemplazar Skills
			// Busca desde "### Skills Reference" hasta el siguiente "### " o fin de archivo
			skillsSectionRegex := regexp.MustCompile(`(?s)(### Skills Reference\n)(?:.*?)(\n### |$)`)
			newContent = skillsSectionRegex.ReplaceAllString(newContent, "${1}"+skillsBlock.String()+"${2}")

			// Reemplazar Rules
			// Busca desde "### Rules\n" hasta el siguiente "### " o fin de archivo
			rulesSectionRegex := regexp.MustCompile(`(?s)(### Rules\n)(?:.*?)(\n### |$)`)
			newContent = rulesSectionRegex.ReplaceAllString(newContent, "${1}"+rulesBlock.String()+"${2}")

			return os.WriteFile(agentPath, []byte(newContent), 0644)
		}
	}

	// Fallback: Generación desde cero (si no existe o estructura irreconocible)
	projectName := filepath.Base(root)

	var content strings.Builder
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
		Version,
		time.Now().Format("2006-01-02"),
		pType,
		strings.ToUpper(pType),
	)

	if len(skills) > 0 {
		fmt.Fprintf(&content, "\n### Skills Reference\nThe following skills are active for this project.\n\n")

		for _, s := range skills {
			fmt.Fprintf(&content, "- [%s (%s)](%s)\n", s.Name, s.Category, s.LocalPath)
		}
	} else {
		fmt.Fprintf(&content, `
### Skills Reference
⚠️ No skills selected. Run 'kolyn init' again to add skills.
`)
	}

	fmt.Fprintf(&content, "\n### Rules\n")

	ruleCounter = 1

	for _, s := range skills {
		if len(s.Rules) > 0 {
			fmt.Fprintf(&content, "\n#### From %s:\n", s.Name)
			for _, r := range s.Rules {
				fmt.Fprintf(&content, "%d. %s\n", ruleCounter, r)
				ruleCounter++
			}
		}
	}

	fmt.Fprintf(&content, "\n#### General:\n")
	fmt.Fprintf(&content, "%d. **Follow the Skills:** Read the reference files above before writing code.\n", ruleCounter)
	ruleCounter++
	fmt.Fprintf(&content, "%d. **Directory Structure:** Respect the existing project structure.\n", ruleCounter)
	ruleCounter++
	fmt.Fprintf(&content, "%d. **Consistency:** Use the same libraries and patterns defined in the stack.\n", ruleCounter)
	ruleCounter++

	return os.WriteFile(filepath.Join(root, "Agent.md"), []byte(content.String()), 0644)
}

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
