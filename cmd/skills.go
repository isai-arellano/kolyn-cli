package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
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
	Run: func(cmd *cobra.Command, args []string) {
		if err := getSkillsJSON(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var skillsPathsCmd = &cobra.Command{
	Use:   "paths",
	Short: "Retorna solo las rutas de los skills",
	Long:  `Retorna una lista de rutas de skills (una por línea) para facilitar el parsing.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := getSkillsPaths(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var skillsListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista skills y permite ver/editar su contenido",
	Long:  `Muestra lista de skills. Selecciona uno para ver contenido, editar o regresar.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSkillsList()
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

func readInput() string {
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// getSkillDescriptionFromFile extrae la descripción de un skill
func getSkillDescriptionFromFile(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
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

// getSkillsDir obtiene el directorio de skills
func getSkillsDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error obteniendo directorio home: %w", err)
	}
	return filepath.Join(homeDir, ".kolyn", "skills"), nil
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

// getSkillsJSON retorna todas las skills en formato JSON
func getSkillsJSON() error {
	skillDirs, err := getSkillsDirs()
	if err != nil {
		return err
	}

	allSkills := []SkillInfo{}
	totalSkills := 0

	for _, baseDir := range skillDirs {
		if _, err := os.Stat(baseDir); os.IsNotExist(err) {
			continue
		}

		// Caminar recursivamente para encontrar .md
		err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
				// Calcular categoría relativa al baseDir
				relPath, _ := filepath.Rel(baseDir, path)
				category := filepath.Dir(relPath)
				if category == "." {
					category = "root"
				}

				skillName := strings.TrimSuffix(info.Name(), ".md")
				content, _ := os.ReadFile(path)

				allSkills = append(allSkills, SkillInfo{
					Name:        skillName,
					Category:    category,
					Path:        path,
					Description: getSkillDescriptionFromFile(string(content)),
				})
				totalSkills++
			}
			return nil
		})
		if err != nil {
			continue
		}
	}

	result := SkillsJSON{
		TotalSkills: totalSkills,
		SkillsDir:   "combined",
		Skills:      allSkills,
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("error generando JSON: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
}

// getSkillsPaths retorna solo las rutas de las skills
func getSkillsPaths() error {
	skillDirs, err := getSkillsDirs()
	if err != nil {
		return err
	}

	for _, baseDir := range skillDirs {
		if _, err := os.Stat(baseDir); os.IsNotExist(err) {
			continue
		}

		filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
				fmt.Println(path)
			}
			return nil
		})
	}
	return nil
}

// runSkillsList muestra lista interactiva de skills
func runSkillsList() error {
	skillDirs, err := getSkillsDirs()
	if err != nil {
		return err
	}

	var allSkills []SkillInfo

	for _, baseDir := range skillDirs {
		if _, err := os.Stat(baseDir); os.IsNotExist(err) {
			continue
		}

		filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
				relPath, _ := filepath.Rel(baseDir, path)
				category := filepath.Dir(relPath)
				if category == "." {
					category = "root"
				}

				skillName := strings.TrimSuffix(info.Name(), ".md")
				content, _ := os.ReadFile(path)

				allSkills = append(allSkills, SkillInfo{
					Name:        skillName,
					Category:    category,
					Path:        path,
					Description: getSkillDescriptionFromFile(string(content)),
				})
			}
			return nil
		})
	}

	if len(allSkills) == 0 {
		fmt.Println("No hay skills disponibles en ~/.kolyn/skills ni en ~/.kolyn/sources")
		return nil
	}

	for {
		// Mostrar lista de skills
		ui.ShowSection("📚 Skills Disponibles")

		for i, skill := range allSkills {
			ui.WhiteText.Printf("  %d. %s/%s\n", i+1, skill.Category, skill.Name)
			if skill.Description != "" {
				ui.Gray.Printf("     %s\n", skill.Description)
			}
		}
		ui.Gray.Println("  0. Volver")
		fmt.Println()

		fmt.Print("Selecciona un skill (número): ")

		input := readInput()

		if input == "0" {
			return nil
		}

		var selection int
		fmt.Sscan(input, &selection)

		if selection < 1 || selection > len(allSkills) {
			ui.PrintWarning("Selección inválida")
			continue
		}

		selectedSkill := allSkills[selection-1]

		// Mostrar opciones para el skill seleccionado
		if err := showSkillOptions(selectedSkill); err != nil {
			return err
		}
	}
}

// showSkillOptions muestra opciones para un skill específico
func showSkillOptions(skill SkillInfo) error {
	for {
		ui.ShowSection(fmt.Sprintf("📄 %s/%s", skill.Category, skill.Name))
		ui.Gray.Printf("Ruta: %s\n\n", skill.Path)

		ui.WhiteText.Println("  1. Ver contenido (lectura)")
		ui.WhiteText.Println("  2. Editar contenido")
		ui.Gray.Println("  0. Volver a la lista")
		fmt.Println()

		fmt.Print("Selecciona una opción: ")

		input := readInput()

		switch input {
		case "0":
			return nil
		case "1":
			if err := viewSkillContent(skill); err != nil {
				return err
			}
		case "2":
			if err := editSkillContent(skill); err != nil {
				return err
			}
		default:
			ui.PrintWarning("Opción inválida")
		}
	}
}

// viewSkillContent muestra el contenido de un skill (solo lectura)
func viewSkillContent(skill SkillInfo) error {
	content, err := os.ReadFile(skill.Path)
	if err != nil {
		return fmt.Errorf("error leyendo skill: %w", err)
	}

	ui.SeparatorDouble()
	fmt.Println(string(content))
	ui.SeparatorDouble()
	fmt.Println()

	ui.Gray.Println("Presiona Enter para continuar...")
	readInput()

	return nil
}

// editSkillContent permite editar el contenido de un skill
func editSkillContent(skill SkillInfo) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	cmd := exec.Command(editor, skill.Path)
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
func GetAllSkillsPaths() ([]string, error) {
	skillDirs, err := getSkillsDirs()
	if err != nil {
		return nil, err
	}

	var paths []string

	for _, baseDir := range skillDirs {
		if _, err := os.Stat(baseDir); os.IsNotExist(err) {
			continue
		}

		filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
				paths = append(paths, path)
			}
			return nil
		})
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
