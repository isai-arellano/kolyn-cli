package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/isai-arellano/kolyn-cli/cmd/ui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Audita el proyecto contra las reglas definidas en las skills",
	Long: `Escanea las skills instaladas en busca de reglas de validación (Frontmatter)
y verifica si el proyecto actual las cumple.

Reglas soportadas en las skills:
- required_deps: Dependencias que DEBEN estar en package.json
- forbidden_deps: Dependencias que NO deben estar
- files_exist: Archivos clave que deben existir`,
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
	ui.ShowSection("🕵️  Kolyn Check - Auditoría de Proyecto")

	// 1. Cargar package.json del proyecto actual
	pkg, err := loadPackageJSON(".")
	if err != nil {
		ui.PrintWarning("No se encontró package.json. Se omitirán chequeos de dependencias.")
	}

	// 2. Obtener skills disponibles
	skills, err := scanSkills(ctx)
	if err != nil {
		return fmt.Errorf("error leyendo skills: %w", err)
	}

	if len(skills) == 0 {
		ui.PrintWarning("No hay skills instaladas para auditar.")
		return nil
	}

	totalChecks := 0
	passedChecks := 0
	warnings := 0

	// 3. Iterar skills y validar
	for _, skill := range skills {
		rules, err := parseSkillRules(skill.Path)
		if err != nil {
			// Si no tiene frontmatter o es inválido, simplemente lo ignoramos o logueamos verbose
			continue
		}

		// Si el skill no tiene reglas de check, saltar
		if len(rules.Check.RequiredDeps) == 0 && len(rules.Check.ForbiddenDeps) == 0 && len(rules.Check.FilesExist) == 0 {
			continue
		}

		ui.WhiteText.Printf("\nEvaluando Skill: %s/%s\n", skill.Category, skill.Name)
		skillPassed := true

		// Check Required Deps
		if pkg != nil {
			for _, dep := range rules.Check.RequiredDeps {
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

			// Check Forbidden Deps
			for _, dep := range rules.Check.ForbiddenDeps {
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

		// Check Files Exist
		for _, file := range rules.Check.FilesExist {
			totalChecks++
			if _, err := os.Stat(file); os.IsNotExist(err) {
				ui.PrintFail("  ❌ Falta archivo: %s", file)
				skillPassed = false
				warnings++
			} else {
				ui.PrintSuccess("  ✅ Archivo encontrado: %s", file)
				passedChecks++
			}
		}

		if skillPassed {
			// ui.Gray.Println("  ✨ Skill cumple con los estándares")
		}
	}

	ui.Separator()
	fmt.Printf("Resumen: %d verificaciones, %d pasadas, %d alertas\n", totalChecks, passedChecks, warnings)

	if warnings > 0 {
		return fmt.Errorf("se encontraron %d problemas en la auditoría", warnings)
	}

	return nil
}

func parseSkillRules(path string) (*SkillRules, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Extraer bloque YAML entre --- y ---
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
