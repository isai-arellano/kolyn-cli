package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/isai-arellano/kolyn-cli/cmd/config"
	"github.com/isai-arellano/kolyn-cli/cmd/ui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Audita el proyecto contra las reglas definidas en las skills",
	Long: `Escanea las skills instaladas en busca de reglas de validación (Frontmatter)
y verifica si el proyecto actual las cumple.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCheck(cmd.Context())
	},
}

// SkillRules define la estructura del Frontmatter en los Markdowns
type SkillRules struct {
	Check struct {
		RequiredDeps  []string `yaml:"required_deps"`
		ForbiddenDeps []string `yaml:"forbidden_deps"`
		FilesExist    []string `yaml:"files_exist"`
	} `yaml:"check"`
}

// PackageJSON estructura mínima para leer dependencias
type PackageJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

func runCheck(ctx context.Context) error {
	// 1. Cargar idioma desde config global
	globalCfg, _ := config.LoadGlobalConfig()
	if globalCfg != nil {
		ui.CurrentLanguage = globalCfg.Language
	}

	ui.ShowSection(ui.GetText("check_start"))

	// 2. Cargar package.json del proyecto actual
	pkg, err := loadPackageJSON(".")
	if err != nil {
		ui.PrintWarning(ui.GetText("no_package_json"))
	}

	// 3. Obtener skills disponibles
	skills, err := scanSkills(ctx)
	if err != nil {
		return fmt.Errorf("error leyendo skills: %w", err)
	}

	if len(skills) == 0 {
		ui.PrintWarning(ui.GetText("no_skills"))
		return nil
	}

	totalChecks := 0
	passedChecks := 0
	warnings := 0

	// 4. Iterar skills y validar
	for _, skill := range skills {
		rules, err := parseSkillRules(skill.Path)
		if err != nil {
			continue
		}

		if len(rules.Check.RequiredDeps) == 0 && len(rules.Check.ForbiddenDeps) == 0 && len(rules.Check.FilesExist) == 0 {
			continue
		}

		ui.WhiteText.Printf(ui.GetText("evaluating_skill", skill.Category, skill.Name))
		skillPassed := true

		// Check Required Deps
		if pkg != nil {
			for _, dep := range rules.Check.RequiredDeps {
				totalChecks++
				if !hasDependency(pkg, dep) {
					ui.PrintFail(ui.GetText("missing_dep", dep))
					skillPassed = false
					warnings++
				} else {
					ui.PrintSuccess(ui.GetText("found_dep", dep))
					passedChecks++
				}
			}

			// Check Forbidden Deps
			for _, dep := range rules.Check.ForbiddenDeps {
				totalChecks++
				if hasDependency(pkg, dep) {
					ui.PrintFail(ui.GetText("forbidden_dep", dep))
					skillPassed = false
					warnings++
				} else {
					passedChecks++
				}
			}
		}

		// Check Files Exist
		for _, file := range rules.Check.FilesExist {
			totalChecks++
			if _, err := os.Stat(file); os.IsNotExist(err) {
				ui.PrintFail(ui.GetText("missing_file", file))
				skillPassed = false
				warnings++
			} else {
				ui.PrintSuccess(ui.GetText("found_file", file))
				passedChecks++
			}
		}

		if skillPassed {
			// ui.Gray.Println("  ✨ Skill OK")
		}
	}

	ui.Separator()
	fmt.Println(ui.GetText("audit_summary", totalChecks, passedChecks, warnings))

	if warnings > 0 {
		return fmt.Errorf(ui.GetText("audit_issues", warnings))
	}

	return nil
}

func parseSkillRules(path string) (*SkillRules, error) {
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
	var rules SkillRules
	if err := yaml.Unmarshal(yamlContent, &rules); err != nil {
		return nil, err
	}

	return &rules, nil
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
