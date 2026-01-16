package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourusername/kolyn/cmd/ui"
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

// getSkillsDir obtiene el directorio de skills
func getSkillsDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error obteniendo directorio home: %w", err)
	}
	return filepath.Join(homeDir, ".kolyn", "skills"), nil
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

// getSkillsJSON retorna todas las skills en formato JSON
func getSkillsJSON() error {
	skillsDir, err := getSkillsDir()
	if err != nil {
		return err
	}

	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		result := SkillsJSON{
			TotalSkills: 0,
			SkillsDir:   skillsDir,
			Skills:      []SkillInfo{},
		}
		jsonData, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(jsonData))
		return nil
	}

	categories, err := os.ReadDir(skillsDir)
	if err != nil {
		return fmt.Errorf("error leyendo directorio de skills: %w", err)
	}

	skills := []SkillInfo{}
	totalSkills := 0

	for _, category := range categories {
		if !category.IsDir() {
			continue
		}

		categoryPath := filepath.Join(skillsDir, category.Name())
		files, err := os.ReadDir(categoryPath)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
				continue
			}

			skillName := strings.TrimSuffix(file.Name(), ".md")
			fullPath := filepath.Join(categoryPath, file.Name())

			content, err := os.ReadFile(fullPath)
			description := ""
			if err == nil {
				description = getSkillDescriptionFromFile(string(content))
			}

			skills = append(skills, SkillInfo{
				Name:        skillName,
				Category:    category.Name(),
				Path:        fullPath,
				Description: description,
			})

			totalSkills++
		}
	}

	result := SkillsJSON{
		TotalSkills: totalSkills,
		SkillsDir:   skillsDir,
		Skills:      skills,
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
	skillsDir, err := getSkillsDir()
	if err != nil {
		return err
	}

	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		return nil
	}

	categories, err := os.ReadDir(skillsDir)
	if err != nil {
		return fmt.Errorf("error leyendo directorio de skills: %w", err)
	}

	for _, category := range categories {
		if !category.IsDir() {
			continue
		}

		categoryPath := filepath.Join(skillsDir, category.Name())
		files, err := os.ReadDir(categoryPath)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
				continue
			}

			fullPath := filepath.Join(categoryPath, file.Name())
			fmt.Println(fullPath)
		}
	}

	return nil
}

// runSkillsList muestra lista interactiva de skills
func runSkillsList() error {
	skillsDir, err := getSkillsDir()
	if err != nil {
		return err
	}

	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		fmt.Println("No se encontró el directorio de skills:", skillsDir)
		return nil
	}

	// Recopilar todas las skills
	var allSkills []SkillInfo
	categories, _ := os.ReadDir(skillsDir)

	for _, category := range categories {
		if !category.IsDir() {
			continue
		}

		categoryPath := filepath.Join(skillsDir, category.Name())
		files, _ := os.ReadDir(categoryPath)

		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
				continue
			}

			skillName := strings.TrimSuffix(file.Name(), ".md")
			fullPath := filepath.Join(categoryPath, file.Name())
			content, _ := os.ReadFile(fullPath)

			allSkills = append(allSkills, SkillInfo{
				Name:        skillName,
				Category:    category.Name(),
				Path:        fullPath,
				Description: getSkillDescriptionFromFile(string(content)),
			})
		}
	}

	if len(allSkills) == 0 {
		fmt.Println("No hay skills disponibles")
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
	skillsDir, err := getSkillsDir()
	if err != nil {
		return nil, err
	}

	var paths []string

	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		return paths, nil
	}

	categories, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, fmt.Errorf("error leyendo directorio de skills: %w", err)
	}

	for _, category := range categories {
		if !category.IsDir() {
			continue
		}

		categoryPath := filepath.Join(skillsDir, category.Name())
		files, err := os.ReadDir(categoryPath)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
				continue
			}

			fullPath := filepath.Join(categoryPath, file.Name())
			paths = append(paths, fullPath)
		}
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
