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

	// 1. Detectar versión actual y sistema
	ui.PrintStep("Buscando actualizaciones...")

	// URL del script de instalación oficial
	installScriptURL := "https://raw.githubusercontent.com/isai-arellano/kolyn-cli/main/install.sh"

	// 2. Descargar y ejecutar el script de instalación
	// Esta es la forma más segura y multiplataforma, ya que el script maneja
	// la detección de arquitectura, descarga del binario correcto y permisos.

	ui.PrintStep("Descargando instalador...")
	resp, err := http.Get(installScriptURL)
	if err != nil {
		return fmt.Errorf("error conectando con GitHub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("error descargando script (status: %d)", resp.StatusCode)
	}

	// Guardar script temporal
	tmpScript, err := os.CreateTemp("", "kolyn-install-*.sh")
	if err != nil {
		return fmt.Errorf("error creando archivo temporal: %w", err)
	}
	defer os.Remove(tmpScript.Name())

	if _, err := io.Copy(tmpScript, resp.Body); err != nil {
		return fmt.Errorf("error guardando script: %w", err)
	}
	tmpScript.Close()

	// Dar permisos de ejecución
	if err := os.Chmod(tmpScript.Name(), 0755); err != nil {
		return fmt.Errorf("error dando permisos: %w", err)
	}

	ui.PrintStep("Instalando nueva versión...")

	// Ejecutar script
	cmd := exec.Command("/bin/sh", tmpScript.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// En Windows podríamos necesitar powershell, pero por ahora asumimos sh/bash (Linux/Mac)
	if runtime.GOOS == "windows" {
		ui.PrintWarning("La actualización automática en Windows requiere Git Bash o WSL.")
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error durante la instalación: %w", err)
	}

	ui.Separator()
	ui.PrintSuccess("¡Kolyn se ha actualizado correctamente!")

	// Verificar nueva versión
	verifyCmd := exec.Command("kolyn", "version")
	output, _ := verifyCmd.CombinedOutput()
	if len(output) > 0 {
		fmt.Println(strings.TrimSpace(string(output)))
	}

	return nil
}
