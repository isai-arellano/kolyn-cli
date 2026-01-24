package cmd

import (
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
	config, err := loadProjectConfig()
	if err != nil {
		if os.IsNotExist(err) {
			ui.PrintWarning("No se encontró .kolyn.json en el directorio actual.")
			return nil
		}
		return fmt.Errorf("error leyendo config: %w", err)
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

func loadProjectConfig() (*KolynConfig, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo directorio actual: %w", err)
	}

	configPath := filepath.Join(cwd, ".kolyn.json")

	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err // os.IsNotExist is handled by caller
	}

	var config KolynConfig
	if err := json.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("error parseando JSON de configuración: %w", err)
	}

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
