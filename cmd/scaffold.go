package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/isai-arellano/kolyn-cli/cmd/ui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var scaffoldCmd = &cobra.Command{
	Use:   "scaffold",
	Short: "Genera o audita la estructura de un proyecto",
	Long:  `Crea nuevos proyectos siguiendo estándares definidos o audita proyectos existentes para asegurar que cumplan con la estructura requerida.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScaffold(cmd.Context())
	},
}

type ScaffoldFile struct {
	Path    string `yaml:"path"`
	Content string `yaml:"content"`
}

type ScaffoldStructure struct {
	Directories []string       `yaml:"directories"`
	Files       []ScaffoldFile `yaml:"files"`
}

type ScaffoldSkill struct {
	Type          string            `yaml:"type"`
	Framework     string            `yaml:"framework"`
	CreateCommand string            `yaml:"create_command"`
	Structure     ScaffoldStructure `yaml:"structure"`
}

func runScaffold(ctx context.Context) error {
	ui.ShowSection("🏗️  Kolyn Scaffold")

	// 1. Select Project Type (Hardcoded for now, can be dynamic later)
	ui.Cyan.Println("Selecciona el tipo de proyecto:")
	fmt.Println("  1. Web (Next.js)")
	fmt.Println("  2. Mobile (Flutter) [Coming Soon]")
	fmt.Println("  3. Backend (Go) [Coming Soon]")
	fmt.Println("  0. Cancelar")

	choice := ui.ReadInput("> ")

	var skillPath string

	switch choice {
	case "1":
		// Locate the skill file
		home, _ := os.UserHomeDir()
		skillPath = filepath.Join(home, ".kolyn", "sources", "github.com-isai-arellano-kolyn-skills", "scaffold", "web", "nextjs.md")

		// Fallback to old path if not found in sources
		if _, err := os.Stat(skillPath); os.IsNotExist(err) {
			// Try without github prefix if sources are flat? No, stick to known structure or fail.
			// Actually, fallback to local repo if developing
			cwd, _ := os.Getwd()
			localPath := filepath.Join(cwd, "skills", "scaffold", "web", "nextjs.md")
			if _, err := os.Stat(localPath); err == nil {
				skillPath = localPath
			} else {
				// Try ~/.kolyn/skills old fallback
				skillPath = filepath.Join(home, ".kolyn", "skills", "scaffold", "web", "nextjs.md")
				if _, err := os.Stat(skillPath); os.IsNotExist(err) {
					return fmt.Errorf("no se encontró la definición de scaffold para Next.js. Ejecuta 'kolyn sync' primero.")
				}
			}
		}
	case "0":
		return nil
	default:
		ui.PrintWarning("Opción no válida o no implementada aún.")
		return nil
	}

	// 2. Parse Skill
	scaffold, err := loadScaffoldSkill(skillPath)
	if err != nil {
		return fmt.Errorf("error leyendo skill: %w", err)
	}

	// 3. Select Mode
	fmt.Println()
	ui.Cyan.Println("¿Qué deseas hacer?")
	fmt.Println("  1. Crear nuevo proyecto")
	fmt.Println("  2. Auditar/Corregir proyecto existente")
	fmt.Println("  0. Cancelar")

	mode := ui.ReadInput("> ")

	switch mode {
	case "1":
		return createNewProject(ctx, scaffold)
	case "2":
		return auditExistingProject(ctx, scaffold)
	case "0":
		return nil
	default:
		ui.PrintWarning("Opción inválida")
		return nil
	}
}

func loadScaffoldSkill(path string) (*ScaffoldSkill, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if !bytes.HasPrefix(content, []byte("---")) {
		return nil, fmt.Errorf("formato inválido: falta frontmatter")
	}

	parts := bytes.SplitN(content, []byte("---"), 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("formato inválido: frontmatter mal formado")
	}

	var skill ScaffoldSkill
	if err := yaml.Unmarshal(parts[1], &skill); err != nil {
		return nil, err
	}

	return &skill, nil
}

func createNewProject(ctx context.Context, scaffold *ScaffoldSkill) error {
	ui.PrintQuestion("Nombre del proyecto:")
	name := ui.ReadInput("> ")
	if name == "" {
		return fmt.Errorf("nombre inválido")
	}

	// 1. Execute Create Command
	cmdStr := strings.ReplaceAll(scaffold.CreateCommand, "{name}", name)
	parts := strings.Fields(cmdStr)

	ui.PrintStep("Ejecutando: %s", cmdStr)

	cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error creando proyecto: %w", err)
	}

	ui.PrintSuccess("Proyecto base creado.")

	// 2. Apply Structure
	projectPath := filepath.Join(".", name)
	if err := applyStructure(projectPath, scaffold.Structure, false); err != nil {
		return err
	}

	ui.Separator()
	ui.PrintSuccess("¡Proyecto '%s' listo!", name)

	// 3. Auto-Init (Generate Agent.md)
	ui.PrintStep("Generando contexto de IA (Agent.md)...")

	// Determine absolute path
	absPath, _ := filepath.Abs(projectPath)

	// We run interactive init so user can confirm features
	// OR we run non-interactive with defaults.
	// Let's ask the user if they want to configure it now
	if ui.AskYesNo("¿Deseas configurar las capabilities del proyecto ahora (DB, Auth, etc)?") {
		if err := RunInitProject(ctx, absPath, true); err != nil {
			ui.PrintWarning("No se pudo completar la inicialización: %v", err)
		}
	} else {
		// Run non-interactive with defaults
		RunInitProject(ctx, absPath, false)
	}

	ui.Gray.Printf("\n  cd %s\n", name)
	ui.Gray.Println("  kolyn check")

	return nil
}

func auditExistingProject(ctx context.Context, scaffold *ScaffoldSkill) error {
	cwd, _ := os.Getwd()
	ui.PrintStep("Analizando estructura en %s...", cwd)

	return applyStructure(cwd, scaffold.Structure, true)
}

func applyStructure(basePath string, structure ScaffoldStructure, interactive bool) error {
	// Dirs
	for _, dir := range structure.Directories {
		dirPath := filepath.Join(basePath, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			if interactive {
				ui.PrintFail("Falta directorio: %s", dir)
				if ui.AskYesNo(fmt.Sprintf("¿Crear %s?", dir)) {
					if err := os.MkdirAll(dirPath, 0755); err != nil {
						ui.PrintError("Error creando dir: %v", err)
					} else {
						ui.PrintSuccess("Creado: %s", dir)
					}
				}
			} else {
				// Auto-create in new project mode
				os.MkdirAll(dirPath, 0755)
				ui.PrintSuccess("Creado: %s", dir)
			}
		} else if interactive {
			ui.PrintSuccess("Existe: %s", dir)
		}
	}

	// Files
	for _, file := range structure.Files {
		filePath := filepath.Join(basePath, file.Path)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			if interactive {
				ui.PrintFail("Falta archivo: %s", file.Path)
				if ui.AskYesNo(fmt.Sprintf("¿Crear %s?", file.Path)) {
					// Ensure parent dir exists
					os.MkdirAll(filepath.Dir(filePath), 0755)
					if err := os.WriteFile(filePath, []byte(file.Content), 0644); err != nil {
						ui.PrintError("Error escribiendo archivo: %v", err)
					} else {
						ui.PrintSuccess("Creado: %s", file.Path)
					}
				}
			} else {
				// Auto-create
				os.MkdirAll(filepath.Dir(filePath), 0755)
				os.WriteFile(filePath, []byte(file.Content), 0644)
				ui.PrintSuccess("Creado: %s", file.Path)
			}
		} else if interactive {
			ui.PrintSuccess("Existe: %s", file.Path)
		}
	}

	return nil
}
