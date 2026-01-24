package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/isai-arellano/kolyn-cli/cmd/ui"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Actualiza kolyn a la última versión disponible",
	Long:  `Descarga e instala la última versión estable de kolyn desde GitHub Releases.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUpdate()
	},
}

func runUpdate() error {
	ui.ShowSection("🔄 Kolyn Update")

	ui.PrintStep("Buscando actualizaciones...")

	if runtime.GOOS == "windows" {
		return runUpdateWindows()
	}

	// URL del script de instalación oficial (Linux/Mac)
	installScriptURL := "https://raw.githubusercontent.com/isai-arellano/kolyn-cli/main/install.sh"

	return downloadAndRunScript(installScriptURL, "/bin/sh")
}

func runUpdateWindows() error {
	// URL del script de instalación oficial (Windows)
	installScriptURL := "https://raw.githubusercontent.com/isai-arellano/kolyn-cli/main/install.ps1"

	// En Windows usamos powershell
	return downloadAndRunScript(installScriptURL, "powershell")
}

func downloadAndRunScript(url, shell string) error {
	ui.PrintStep("Descargando instalador...")
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error conectando con GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
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
	defer os.Remove(tmpScript.Name()) // Clean up

	// Copiar contenido
	if _, err := io.Copy(tmpScript, resp.Body); err != nil {
		return fmt.Errorf("error guardando script: %w", err)
	}
	tmpScript.Close()

	// Dar permisos (solo relevante en unix, pero no daña en win)
	os.Chmod(tmpScript.Name(), 0755)

	ui.PrintStep("Instalando nueva versión...")

	// Ejecutar script
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// En Windows necesitamos pasar argumentos específicos para bypass de políticas
		cmd = exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-File", tmpScript.Name())
	} else {
		cmd = exec.Command(shell, tmpScript.Name())
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error durante la instalación: %w", err)
	}

	ui.Separator()
	ui.PrintSuccess("¡Kolyn se ha actualizado correctamente!")

	// Verificar nueva versión (ignorar error si falla por PATH aun no actualizado en la sesión)
	verifyCmd := exec.Command("kolyn", "version")
	output, _ := verifyCmd.CombinedOutput()
	if len(output) > 0 {
		fmt.Println(strings.TrimSpace(string(output)))
	} else if runtime.GOOS == "windows" {
		ui.PrintInfo("Reinicia tu terminal para usar la nueva versión.")
	}

	return nil
}
