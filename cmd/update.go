package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/isai-arellano/kolyn-cli/cmd/ui"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Actualiza kolyn a la última versión disponible",
	Long:  `Descarga e instala la última versión estable de kolyn desde GitHub Releases.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUpdate(cmd.Context())
	},
}

func runUpdate(ctx context.Context) error {
	ui.ShowSection("🔄 Kolyn Update")

	ui.PrintStep("Buscando actualizaciones...")

	if runtime.GOOS == "windows" {
		return runUpdateWindows(ctx)
	}

	// URL del script de instalación oficial (Linux/Mac)
	installScriptURL := "https://raw.githubusercontent.com/isai-arellano/kolyn-cli/main/install.sh"

	return downloadAndRunScript(ctx, installScriptURL, "/bin/sh")
}

func runUpdateWindows(ctx context.Context) error {
	// URL del script de instalación oficial (Windows)
	installScriptURL := "https://raw.githubusercontent.com/isai-arellano/kolyn-cli/main/install.ps1"

	// En Windows usamos powershell
	return downloadAndRunScript(ctx, installScriptURL, "powershell")
}

func downloadAndRunScript(ctx context.Context, url, shell string) error {
	ui.PrintStep("Descargando instalador...")

	// Use Context for HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("error creando petición HTTP: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error conectando con GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error descargando script (status: %d)", resp.StatusCode)
	}

	// Extension depende del OS
	ext := ".sh"
	if runtime.GOOS == "windows" {
		ext = ".ps1"
	}

	// Guardar script temporal
	tmpScript, err := os.CreateTemp("", "kolyn-install-*"+ext)
	if err != nil {
		return fmt.Errorf("error creando archivo temporal: %w", err)
	}
	defer func() {
		if err := os.Remove(tmpScript.Name()); err != nil {
			// Log error if cleanup fails, but don't fail the operation
			fmt.Fprintf(os.Stderr, "warning: failed to cleanup temp file: %v\n", err)
		}
	}()

	// Copiar contenido
	if _, err := io.Copy(tmpScript, resp.Body); err != nil {
		return fmt.Errorf("error guardando script: %w", err)
	}
	// Close explicitly to ensure flush
	if err := tmpScript.Close(); err != nil {
		return fmt.Errorf("error cerrando archivo temporal: %w", err)
	}

	// Dar permisos (solo relevante en unix, pero no daña en win)
	if err := os.Chmod(tmpScript.Name(), 0755); err != nil {
		return fmt.Errorf("error asignando permisos de ejecución: %w", err)
	}

	ui.PrintStep("Instalando nueva versión...")

	// Ejecutar script
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// En Windows necesitamos pasar argumentos específicos para bypass de políticas
		cmd = exec.CommandContext(ctx, "powershell", "-ExecutionPolicy", "Bypass", "-File", tmpScript.Name())
	} else {
		cmd = exec.CommandContext(ctx, shell, tmpScript.Name())
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error durante la instalación: %w", err)
	}

	ui.Separator()
	ui.PrintSuccess("¡Kolyn se ha actualizado correctamente!")

	// Verificar nueva versión (ignorar error si falla por PATH aun no actualizado en la sesión)
	verifyCmd := exec.CommandContext(ctx, "kolyn", "version")
	output, err := verifyCmd.CombinedOutput()
	if err == nil && len(output) > 0 {
		fmt.Println(strings.TrimSpace(string(output)))
	} else if runtime.GOOS == "windows" {
		ui.PrintInfo("Reinicia tu terminal para usar la nueva versión.")
	}

	return nil
}
