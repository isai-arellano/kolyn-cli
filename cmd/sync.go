package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/isai-arellano/kolyn-cli/cmd/ui"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sincroniza skills desde fuentes externas definidas en .kolyn.json",
	Long: `Descarga y actualiza repositorios de skills definidos en el archivo .kolyn.json del proyecto actual.
	
Ejemplo de .kolyn.json:
{
  "project_name": "mi-proyecto",
  "skills_sources": [
    "https://github.com/mi-org/backend-standards.git"
  ]
}`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSyncCommand(cmd.Context())
	},
}

type KolynConfig struct {
	ProjectName   string   `json:"project_name"`
	SkillsSources []string `json:"skills_sources"`
}

func runSyncCommand(ctx context.Context) error {
	// 1. Leer configuración
	config, err := loadProjectConfig(ctx)
	if err != nil {
		return err
	}

	if config == nil {
		// User cancelled creation
		return nil
	}

	ui.ShowSection(fmt.Sprintf("🔄 Sincronizando Skills para: %s", config.ProjectName))

	// 2. Preparar directorio de sources
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error obteniendo directorio home: %w", err)
	}

	sourcesBaseDir := filepath.Join(homeDir, ".kolyn", "sources")
	if err := os.MkdirAll(sourcesBaseDir, 0755); err != nil {
		return fmt.Errorf("error creando directorio de sources: %w", err)
	}

	// 3. Procesar cada fuente
	for _, sourceURL := range config.SkillsSources {
		if err := syncSource(ctx, sourceURL, sourcesBaseDir); err != nil {
			ui.PrintError("Fallo al sincronizar %s: %v", sourceURL, err)
		}
	}

	ui.PrintSuccess("Sincronización completada!")
	return nil
}

func loadProjectConfig(ctx context.Context) (*KolynConfig, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo directorio actual: %w", err)
	}

	configPath := filepath.Join(cwd, ".kolyn.json")

	content, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return promptCreateConfig(ctx, configPath)
		}
		return nil, fmt.Errorf("error leyendo config: %w", err) // wrap unknown errors
	}

	var config KolynConfig
	if err := json.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("error parseando JSON de configuración: %w", err)
	}

	return &config, nil
}

func promptCreateConfig(ctx context.Context, configPath string) (*KolynConfig, error) {
	ui.PrintWarning("No se encontró .kolyn.json en el directorio actual.")
	fmt.Println()
	ui.YellowText.Println("¿Deseas crear un archivo de configuración ahora? [s/N]: ")
	fmt.Print("> ")

	reader := bufio.NewReader(os.Stdin)
	answer, err := readInput(reader)
	if err != nil {
		return nil, fmt.Errorf("error leyendo entrada: %w", err)
	}

	if strings.ToLower(answer) != "s" && strings.ToLower(answer) != "si" && strings.ToLower(answer) != "yes" && strings.ToLower(answer) != "y" {
		ui.PrintInfo("Operación cancelada. Crea un archivo .kolyn.json manualmente para continuar.")
		return nil, nil // return nil config to signal cancellation without error
	}

	// 1. Project Name
	cwd, _ := os.Getwd()
	defaultName := filepath.Base(cwd)

	fmt.Printf("Nombre del proyecto [%s]: ", defaultName)
	projectName, err := readInput(reader)
	if err != nil {
		return nil, fmt.Errorf("error leyendo entrada: %w", err)
	}
	if projectName == "" {
		projectName = defaultName
	}

	// 2. Repo URLs
	var sources []string
	ui.Gray.Println("Ingresa las URLs de los repositorios de skills (una por línea).")
	ui.Gray.Println("Deja la línea vacía y presiona Enter para terminar.")

	for {
		fmt.Print("Repo URL: ")
		url, err := readInput(reader)
		if err != nil {
			return nil, fmt.Errorf("error leyendo entrada: %w", err)
		}
		if url == "" {
			break
		}
		sources = append(sources, url)
	}

	if len(sources) == 0 {
		ui.PrintWarning("No se ingresaron repositorios. El archivo se creará sin sources.")
	}

	config := KolynConfig{
		ProjectName:   projectName,
		SkillsSources: sources,
	}

	// Write file
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error generando JSON: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return nil, fmt.Errorf("error escribiendo .kolyn.json: %w", err)
	}

	ui.PrintSuccess("Archivo .kolyn.json creado exitosamente!")
	return &config, nil
}

func syncSource(ctx context.Context, url, baseDir string) error {
	folderName := sanitizeRepoName(url)
	targetDir := filepath.Join(baseDir, folderName)

	if _, err := os.Stat(targetDir); err == nil {
		// Existe: hacer pull
		ui.PrintStep("Actualizando %s...", folderName)
		cmd := exec.CommandContext(ctx, "git", "pull")
		cmd.Dir = targetDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git pull falló: %s (%w)", string(output), err)
		}
	} else {
		// No existe: hacer clone
		ui.PrintStep("Descargando %s...", url)
		cmd := exec.CommandContext(ctx, "git", "clone", url, targetDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git clone falló: %s (%w)", string(output), err)
		}
	}

	return nil
}

func sanitizeRepoName(url string) string {
	// Limpieza básica: quitar protocolo, .git y reemplazar barras
	name := url
	name = strings.TrimPrefix(name, "https://")
	name = strings.TrimPrefix(name, "http://")
	name = strings.TrimPrefix(name, "git@")
	name = strings.TrimSuffix(name, ".git")
	name = strings.ReplaceAll(name, ":", "/")
	name = strings.ReplaceAll(name, "/", "-")
	return name
}
