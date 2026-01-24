package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/isai-arellano/kolyn-cli/cmd/ui"
	"github.com/spf13/cobra"
)

var skillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Retorna las rutas de los skills disponibles para contexto de IA",
	Long:  `Retorna un JSON con todas las skills disponibles y sus rutas. La IA puede usar estas rutas para leer el contenido de cada skill.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSkillsJSON(cmd.Context())
	},
}

var skillsPathsCmd = &cobra.Command{
	Use:   "paths",
	Short: "Retorna solo las rutas de los skills",
	Long:  `Retorna una lista de rutas de skills (una por línea) para facilitar el parsing.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSkillsPaths(cmd.Context())
	},
}

var skillsListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista skills y permite ver/editar su contenido",
	Long:  `Muestra lista de skills. Selecciona uno para ver contenido, editar o regresar.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSkillsList(cmd.Context())
	},
}

// SkillInfo representa la información de un skill
type SkillInfo struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Path        string `json:"path"`
	Description string `json:"description,omitempty"`
}

// SkillsJSON estructura para retornar todas las skills
type SkillsJSON struct {
	TotalSkills int         `json:"total_skills"`
	SkillsDir   string      `json:"skills_dir"`
	Skills      []SkillInfo `json:"skills"`
}

// getSkillDescriptionFromFile extrae la descripción de un skill
func getSkillDescriptionFromFile(content string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "- **Description**") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
		if strings.HasPrefix(line, "**Description**") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

// getSkillsDirs obtiene todos los directorios donde buscar skills (local y sources)
func getSkillsDirs() ([]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo directorio home: %w", err)
	}

	dirs := []string{}

	// 1. Skills locales (~/.kolyn/skills)
	localSkills := filepath.Join(homeDir, ".kolyn", "skills")
	dirs = append(dirs, localSkills)

	// 2. Skills sincronizados (~/.kolyn/sources/*)
	sourcesDir := filepath.Join(homeDir, ".kolyn", "sources")
	if entries, err := os.ReadDir(sourcesDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				dirs = append(dirs, filepath.Join(sourcesDir, entry.Name()))
			}
		}
	}

	return dirs, nil
}

// scanSkills busca todos los skills disponibles
func scanSkills(ctx context.Context) ([]SkillInfo, error) {
	skillDirs, err := getSkillsDirs()
	if err != nil {
		return nil, err
	}

	var allSkills []SkillInfo

	for _, baseDir := range skillDirs {
		if _, err := os.Stat(baseDir); os.IsNotExist(err) {
			continue
		}

		err := filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // Skip errors accessing files
			}

			// Check context cancellation
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if !d.IsDir() && strings.HasSuffix(d.Name(), ".md") {
				// Calcular categoría relativa al baseDir
				relPath, _ := filepath.Rel(baseDir, path)
				category := filepath.Dir(relPath)
				if category == "." {
					category = "root"
				}

				skillName := strings.TrimSuffix(d.Name(), ".md")
				contentBytes, err := os.ReadFile(path)
				if err != nil {
					return nil // Skip unreadable files
				}

				allSkills = append(allSkills, SkillInfo{
					Name:        skillName,
					Category:    category,
					Path:        path,
					Description: getSkillDescriptionFromFile(string(contentBytes)),
				})
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error escaneando directorio %s: %w", baseDir, err)
		}
	}

	return allSkills, nil
}

// runSkillsJSON retorna todas las skills en formato JSON
func runSkillsJSON(ctx context.Context) error {
	skills, err := scanSkills(ctx)
	if err != nil {
		return err
	}

	result := SkillsJSON{
		TotalSkills: len(skills),
		SkillsDir:   "combined",
		Skills:      skills,
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("error generando JSON: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
}

// runSkillsPaths retorna solo las rutas de las skills
func runSkillsPaths(ctx context.Context) error {
	skills, err := scanSkills(ctx)
	if err != nil {
		return err
	}

	for _, skill := range skills {
		fmt.Println(skill.Path)
	}
	return nil
}

// runSkillsList muestra lista interactiva de skills
func runSkillsList(ctx context.Context) error {
	skills, err := scanSkills(ctx)
	if err != nil {
		return err
	}

	if len(skills) == 0 {
		ui.PrintWarning("No hay skills disponibles en ~/.kolyn/skills ni en ~/.kolyn/sources")
		return nil
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		// Mostrar lista de skills
		ui.ShowSection("📚 Skills Disponibles")

		for i, skill := range skills {
			ui.WhiteText.Printf("  %d. %s/%s\n", i+1, skill.Category, skill.Name)
			if skill.Description != "" {
				ui.Gray.Printf("     %s\n", skill.Description)
			}
		}
		ui.Gray.Println("  0. Volver")
		fmt.Println()

		fmt.Print("Selecciona un skill (número): ")

		input, err := readInput(reader)
		if err != nil {
			return err
		}

		if input == "0" {
			return nil
		}

		var selection int
		if _, err := fmt.Sscan(input, &selection); err != nil {
			ui.PrintWarning("Entrada inválida")
			continue
		}

		if selection < 1 || selection > len(skills) {
			ui.PrintWarning("Selección inválida")
			continue
		}

		selectedSkill := skills[selection-1]

		// Mostrar opciones para el skill seleccionado
		if err := showSkillOptions(ctx, reader, selectedSkill); err != nil {
			return err
		}
	}
}

// showSkillOptions muestra opciones para un skill específico
func showSkillOptions(ctx context.Context, reader *bufio.Reader, skill SkillInfo) error {
	for {
		ui.ShowSection(fmt.Sprintf("📄 %s/%s", skill.Category, skill.Name))
		ui.Gray.Printf("Ruta: %s\n\n", skill.Path)

		ui.WhiteText.Println("  1. Ver contenido (lectura)")
		ui.WhiteText.Println("  2. Editar contenido")
		ui.Gray.Println("  0. Volver a la lista")
		fmt.Println()

		fmt.Print("Selecciona una opción: ")

		input, err := readInput(reader)
		if err != nil {
			return err
		}

		switch input {
		case "0":
			return nil
		case "1":
			if err := viewSkillContent(reader, skill); err != nil {
				return err
			}
		case "2":
			if err := editSkillContent(ctx, skill); err != nil {
				return err
			}
		default:
			ui.PrintWarning("Opción inválida")
		}
	}
}

// viewSkillContent muestra el contenido de un skill (solo lectura)
func viewSkillContent(reader *bufio.Reader, skill SkillInfo) error {
	content, err := os.ReadFile(skill.Path)
	if err != nil {
		return fmt.Errorf("error leyendo skill: %w", err)
	}

	ui.SeparatorDouble()
	fmt.Println(string(content))
	ui.SeparatorDouble()
	fmt.Println()

	ui.Gray.Println("Presiona Enter para continuar...")
	_, _ = readInput(reader)

	return nil
}

// editSkillContent permite editar el contenido de un skill
func editSkillContent(ctx context.Context, skill SkillInfo) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	cmd := exec.CommandContext(ctx, editor, skill.Path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	ui.Gray.Printf("Editando %s con %s\n", skill.Path, editor)
	ui.Gray.Println("Guarda y sale para continuar...")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error editando skill: %w", err)
	}

	ui.PrintSuccess("Skill actualizado!")
	return nil
}

// GetAllSkillsPaths retorna todas las rutas de skills (para uso interno)
// Deprecated: use scanSkills instead or if needed for external consumers, refactor to use context
func GetAllSkillsPaths() ([]string, error) {
	// Simple wrapper for backward compatibility if needed, though scanSkills is preferred
	skills, err := scanSkills(context.Background())
	if err != nil {
		return nil, err
	}
	paths := make([]string, len(skills))
	for i, s := range skills {
		paths[i] = s.Path
	}
	return paths, nil
}

// GetSkillContent lee el contenido de un skill por su ruta
func GetSkillContent(skillPath string) (string, error) {
	content, err := os.ReadFile(skillPath)
	if err != nil {
		return "", fmt.Errorf("error leyendo skill: %w", err)
	}
	return string(content), nil
}
