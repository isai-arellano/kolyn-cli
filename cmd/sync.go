package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/isai-arellano/kolyn-cli/cmd/config"
	"github.com/isai-arellano/kolyn-cli/cmd/ui"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sincroniza skills (Globales o del Proyecto)",
	Long: `Descarga y actualiza repositorios de skills.
	
Prioridad:
1. Archivo local .kolyn.json (si existe)
2. Configuración global ~/.kolyn/config.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSyncCommand(cmd.Context())
	},
}

type KolynConfig struct {
	ProjectName   string   `json:"project_name"`
	SkillsSources []string `json:"skills_sources"`
}

func runSyncCommand(ctx context.Context) error {
	// 1. Intentar cargar config global para setear idioma (si existe)
	globalCfg, _ := config.LoadGlobalConfig()
	if globalCfg != nil {
		ui.CurrentLanguage = globalCfg.Language
	}

	// 2. Verificar si hay config local (.kolyn.json)
	cwd, _ := os.Getwd()
	localConfigPath := filepath.Join(cwd, ".kolyn.json")
	var sources []string
	var mode string // "local" or "global"

	if _, err := os.Stat(localConfigPath); err == nil {
		// --- MODO LOCAL ---
		ui.PrintInfo(ui.GetText("using_local"))
		localCfg, err := loadLocalConfig(localConfigPath)
		if err != nil {
			return err
		}
		sources = localCfg.SkillsSources
		mode = "local"
	} else {
		// --- MODO GLOBAL ---
		if globalCfg == nil {
			// Primera vez que corre: Setup inicial
			ui.PrintInfo(ui.GetText("no_config"))

			// Seleccionar idioma
			lang := ui.SelectLanguage()
			ui.CurrentLanguage = lang

			// Default sources
			defaultSources := []string{
				"https://github.com/isai-arellano/kolyn-cli.git", // Self-reference for skills
			}

			// Guardar config global
			newGlobal := &config.GlobalConfig{
				Language:      lang,
				SkillsSources: defaultSources,
			}

			if err := config.SaveGlobalConfig(newGlobal); err != nil {
				return fmt.Errorf("error saving global config: %w", err)
			}
			ui.PrintSuccess(ui.GetText("global_created"))
			sources = defaultSources
		} else {
			ui.PrintInfo(ui.GetText("using_global"))
			sources = globalCfg.SkillsSources
		}
		mode = "global"
	}

	ui.ShowSection(ui.GetText("sync_start"))

	// 3. Preparar directorio de sources
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home dir: %w", err)
	}

	sourcesBaseDir := filepath.Join(homeDir, ".kolyn", "sources")
	if err := os.MkdirAll(sourcesBaseDir, 0755); err != nil {
		return fmt.Errorf("error creating sources dir: %w", err)
	}

	// 4. Procesar fuentes
	for _, sourceURL := range sources {
		if err := syncSource(ctx, sourceURL, sourcesBaseDir); err != nil {
			ui.PrintError("Fallo al sincronizar %s: %v", sourceURL, err)
		}
	}

	ui.PrintSuccess(ui.GetText("sync_success"))

	if mode == "global" {
		ui.Gray.Println("\nTip: Si quieres skills específicas para un proyecto, crea un archivo .kolyn.json en la raíz del proyecto.")
	}

	return nil
}

func loadLocalConfig(path string) (*KolynConfig, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg KolynConfig
	if err := json.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing .kolyn.json: %w", err)
	}
	return &cfg, nil
}

func syncSource(ctx context.Context, url, baseDir string) error {
	folderName := sanitizeRepoName(url)
	targetDir := filepath.Join(baseDir, folderName)

	if _, err := os.Stat(targetDir); err == nil {
		// Update
		ui.PrintStep(ui.GetText("updating_skills", folderName))
		cmd := exec.CommandContext(ctx, "git", "pull")
		cmd.Dir = targetDir
		cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0") // Prevent hanging

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		if err != nil {
			if strings.Contains(outputStr, "Permission denied") ||
				strings.Contains(outputStr, "Authentication failed") {
				return fmt.Errorf("\n❌ %s\n%s", ui.GetText("repo_access_error"), outputStr)
			}
			return fmt.Errorf("git pull failed: %s (%w)", outputStr, err)
		}
	} else {
		// Clone
		ui.PrintStep(ui.GetText("installing_skills", url))
		cmd := exec.CommandContext(ctx, "git", "clone", url, targetDir)
		cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0") // Prevent hanging

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		if err != nil {
			if strings.Contains(outputStr, "Permission denied") ||
				strings.Contains(outputStr, "Authentication failed") ||
				strings.Contains(outputStr, "could not read Username") {
				return fmt.Errorf("\n❌ %s\n%s", ui.GetText("repo_access_error"), outputStr)
			}
			return fmt.Errorf("git clone failed: %s (%w)", outputStr, err)
		}
	}
	return nil
}

func sanitizeRepoName(url string) string {
	name := url
	name = strings.TrimPrefix(name, "https://")
	name = strings.TrimPrefix(name, "http://")
	name = strings.TrimPrefix(name, "git@")
	name = strings.TrimSuffix(name, ".git")
	name = strings.ReplaceAll(name, ":", "/")
	name = strings.ReplaceAll(name, "/", "-")
	return name
}
