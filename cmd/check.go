package cmd

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/isai-arellano/kolyn-cli/cmd/config"
	"github.com/isai-arellano/kolyn-cli/cmd/ui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Audita el proyecto usando el contexto de Agent.md",
	Long: `Lee el archivo Agent.md para entender el tipo de proyecto y features activas,
y luego valida que el código cumpla con las reglas definidas en las skills.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCheck(cmd.Context())
	},
}

// SkillFrontmatter define la estructura del Frontmatter en los Markdowns
type SkillFrontmatter struct {
	Name        string     `yaml:"name"`
	Description string     `yaml:"description"`
	AppliesTo   []string   `yaml:"applies_to"`
	Capability  string     `yaml:"capability"`
	Check       SkillCheck `yaml:"check"`
}

type SkillCheck struct {
	RequiredDeps  []string `yaml:"required_deps"`
	DepsExistAny  []string `yaml:"deps_exist_any"`
	ForbiddenDeps []string `yaml:"forbidden_deps"`
	FilesExist    []string `yaml:"files_exist"`
	FilesExistAny []string `yaml:"files_exist_any"`
	EnvVars       []string `yaml:"env_vars"`
	FailMessage   string   `yaml:"fail_message"`
}

// PackageJSON estructura mínima para leer dependencias
type PackageJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

type AgentContext struct {
	ProjectType string
	Features    map[string]bool
}

func runCheck(ctx context.Context) error {
	// 1. Cargar idioma
	globalCfg, _ := config.LoadGlobalConfig()
	if globalCfg != nil {
		ui.CurrentLanguage = globalCfg.Language
	}

	cwd, _ := os.Getwd()
	agentPath := filepath.Join(cwd, "Agent.md")

	// 2. Leer Agent.md
	if _, err := os.Stat(agentPath); os.IsNotExist(err) {
		ui.YellowText.Println("⚠️  No se encontró Agent.md en este proyecto.")
		ui.Gray.Println("   Ejecuta 'kolyn init' para configurar el contexto.")
		return nil
	}

	agentCtx, err := parseAgentContext(agentPath)
	if err != nil {
		return fmt.Errorf("error leyendo Agent.md: %w", err)
	}

	ui.ShowSection("🕵️  Kolyn Check")
	ui.Cyan.Printf("   🔍 Tipo: %s\n", agentCtx.ProjectType)
	featuresList := []string{}
	for k := range agentCtx.Features {
		featuresList = append(featuresList, k)
	}
	ui.Cyan.Printf("   📋 Features: %s\n\n", strings.Join(featuresList, ", "))

	// 3. Cargar package.json (si aplica)
	pkg, _ := loadPackageJSON(cwd)
	if pkg == nil && (agentCtx.ProjectType == "nextjs" || agentCtx.ProjectType == "node") {
		ui.PrintWarning("No se encontró package.json. Se omitirán chequeos de dependencias.")
	}

	// 4. Obtener skills disponibles
	skills, err := scanSkills(ctx)
	if err != nil {
		return fmt.Errorf("error leyendo skills: %w", err)
	}

	totalChecks := 0
	passedChecks := 0
	warnings := 0

	ui.Separator()

	// 5. Validar cada skill
	for _, skill := range skills {
		fm, err := parseSkillFrontmatter(skill.Path)
		if err != nil {
			continue // Skip files without frontmatter
		}

		// Filtro 1: Applies To (Project Type)
		if len(fm.AppliesTo) > 0 && !contains(fm.AppliesTo, agentCtx.ProjectType) {
			continue // No aplica a este tipo de proyecto
		}

		// Filtro 2: Capability (Feature)
		// Si el skill requiere una capability y el proyecto NO la tiene activa, SKIP
		// Excepción: si capability es "core", siempre aplica (si pasó el filtro de tipo)
		if fm.Capability != "" && fm.Capability != "core" {
			if !agentCtx.Features[fm.Capability] {
				ui.Gray.Printf("⏭️  Omitido: %s/%s [%s] (feature inactiva)\n", skill.Category, skill.Name, fm.Capability)
				continue
			}
		}

		// Si no tiene reglas de check, skip
		rules := fm.Check
		if len(rules.RequiredDeps) == 0 && len(rules.ForbiddenDeps) == 0 && len(rules.FilesExist) == 0 &&
			len(rules.DepsExistAny) == 0 && len(rules.FilesExistAny) == 0 && len(rules.EnvVars) == 0 {
			continue
		}

		ui.WhiteText.Printf("📦 Evaluando: %s/%s [%s]\n", skill.Category, skill.Name, fm.Capability)
		skillPassed := true

		// --- CHECKS ---

		// 1. Required Deps
		if pkg != nil {
			for _, dep := range rules.RequiredDeps {
				totalChecks++
				if !hasDependency(pkg, dep) {
					ui.PrintFail("  ❌ Falta dependencia: %s", dep)
					skillPassed = false
					warnings++
				} else {
					ui.PrintSuccess("  ✅ Dependencia encontrada: %s", dep)
					passedChecks++
				}
			}

			// 2. Deps Exist Any
			if len(rules.DepsExistAny) > 0 {
				totalChecks++
				foundAny := false
				for _, dep := range rules.DepsExistAny {
					if hasDependency(pkg, dep) {
						foundAny = true
						ui.PrintSuccess("  ✅ Dependencia encontrada (any): %s", dep)
						passedChecks++
						break
					}
				}
				if !foundAny {
					ui.PrintFail("  ❌ Se requiere al menos una de estas deps: %s", strings.Join(rules.DepsExistAny, ", "))
					skillPassed = false
					warnings++
				}
			}

			// 3. Forbidden Deps
			for _, dep := range rules.ForbiddenDeps {
				totalChecks++
				if hasDependency(pkg, dep) {
					ui.PrintFail("  ❌ Dependencia prohibida detectada: %s", dep)
					skillPassed = false
					warnings++
				} else {
					passedChecks++
				}
			}
		}

		// 4. Files Exist
		for _, file := range rules.FilesExist {
			totalChecks++
			if _, err := os.Stat(filepath.Join(cwd, file)); os.IsNotExist(err) {
				ui.PrintFail("  ❌ Falta archivo: %s", file)
				skillPassed = false
				warnings++
			} else {
				ui.PrintSuccess("  ✅ Archivo encontrado: %s", file)
				passedChecks++
			}
		}

		// 5. Files Exist Any
		if len(rules.FilesExistAny) > 0 {
			totalChecks++
			foundAny := false
			for _, file := range rules.FilesExistAny {
				if _, err := os.Stat(filepath.Join(cwd, file)); err == nil {
					foundAny = true
					ui.PrintSuccess("  ✅ Archivo encontrado (any): %s", file)
					passedChecks++
					break
				}
			}
			if !foundAny {
				ui.PrintFail("  ❌ Se requiere al menos uno de estos archivos: %s", strings.Join(rules.FilesExistAny, ", "))
				skillPassed = false
				warnings++
			}
		}

		// 6. Env Vars (Simple check if .env exists and contains var - naive implementation)
		if len(rules.EnvVars) > 0 {
			// Check .env file content
			envContent, _ := os.ReadFile(filepath.Join(cwd, ".env"))
			envStr := string(envContent)

			for _, v := range rules.EnvVars {
				totalChecks++
				if !strings.Contains(envStr, v+"=") {
					ui.PrintFail("  ❌ Falta Variable de Entorno: %s", v)
					skillPassed = false
					warnings++
				} else {
					ui.PrintSuccess("  ✅ Env Var encontrada: %s", v)
					passedChecks++
				}
			}
		}

		if !skillPassed && rules.FailMessage != "" {
			ui.YellowText.Printf("  💡 Tip: %s\n", rules.FailMessage)
		}
		fmt.Println()
	}

	ui.Separator()
	fmt.Println(ui.GetText("audit_summary", totalChecks, passedChecks, warnings))

	if warnings > 0 {
		return fmt.Errorf("se encontraron %d problemas en la auditoría", warnings)
	}

	return nil
}

func parseAgentContext(path string) (*AgentContext, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	ctx := &AgentContext{
		Features:    make(map[string]bool),
		ProjectType: "generic", // Default
	}

	scanner := bufio.NewScanner(file)
	inFeatures := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Parse Project Type
		if strings.HasPrefix(line, "Project Type:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				ctx.ProjectType = strings.TrimSpace(parts[1])
			}
		}

		// Parse Features Block
		if strings.HasPrefix(line, "## Features") {
			inFeatures = true
			continue
		}
		if inFeatures && strings.HasPrefix(line, "##") {
			inFeatures = false // End of block
		}

		if inFeatures && strings.HasPrefix(line, "-") {
			feature := strings.TrimPrefix(line, "-")
			feature = strings.TrimSpace(feature)
			if feature != "" {
				ctx.Features[feature] = true
			}
		}
	}

	return ctx, nil
}

func parseSkillFrontmatter(path string) (*SkillFrontmatter, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if !bytes.HasPrefix(content, []byte("---")) {
		return nil, fmt.Errorf("no frontmatter")
	}

	parts := bytes.SplitN(content, []byte("---"), 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("frontmatter mal formado")
	}

	yamlContent := parts[1]
	var fm SkillFrontmatter
	if err := yaml.Unmarshal(yamlContent, &fm); err != nil {
		return nil, err
	}

	return &fm, nil
}

func loadPackageJSON(path string) (*PackageJSON, error) {
	file, err := os.ReadFile(filepath.Join(path, "package.json"))
	if err != nil {
		return nil, err
	}

	var pkg PackageJSON
	if err := json.Unmarshal(file, &pkg); err != nil {
		return nil, err
	}
	return &pkg, nil
}

func hasDependency(pkg *PackageJSON, dep string) bool {
	if _, ok := pkg.Dependencies[dep]; ok {
		return true
	}
	if _, ok := pkg.DevDependencies[dep]; ok {
		return true
	}
	return false
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
