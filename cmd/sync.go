package cmd

import (
	"context"
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
	Short: "Sincroniza skills (Globales)",
	Long:  `Descarga y actualiza repositorios de skills definidos en ~/.kolyn/config.json.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSyncCommand(cmd.Context())
	},
}

func runSyncCommand(ctx context.Context) error {
	// 1. Cargar config global
	globalCfg, _ := config.LoadGlobalConfig()
	if globalCfg != nil {
		ui.CurrentLanguage = globalCfg.Language
	}

	var sources []string

	if globalCfg == nil {
		// Primera vez que corre: Setup inicial interactivo
		ui.PrintInfo(ui.GetText("no_config"))

		if err := runConfigInit(ctx); err != nil {
			return err
		}

		// Recargar la configuración recién creada
		globalCfg, err := config.LoadGlobalConfig()
		if err != nil {
			return fmt.Errorf("error reloading global config: %w", err)
		}
		if globalCfg == nil {
			return fmt.Errorf("configuration was not saved correctly")
		}
		sources = globalCfg.SkillsSources
	} else {
		ui.PrintInfo(ui.GetText("using_global"))
		sources = globalCfg.SkillsSources
	}

	ui.ShowSection(ui.GetText("sync_start"))

	// 2. Preparar directorio de sources
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home dir: %w", err)
	}

	sourcesBaseDir := filepath.Join(homeDir, ".kolyn", "sources")
	if err := os.MkdirAll(sourcesBaseDir, 0755); err != nil {
		return fmt.Errorf("error creating sources dir: %w", err)
	}

	// 3. Procesar fuentes
	for _, sourceURL := range sources {
		if err := syncSource(ctx, sourceURL, sourcesBaseDir); err != nil {
			ui.PrintError("Fallo al sincronizar %s: %v", sourceURL, err)
		}
	}

	ui.PrintSuccess(ui.GetText("sync_success"))

	return nil
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
