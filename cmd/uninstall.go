package cmd

import (
	"bufio"
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

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Desinstala Kolyn CLI del sistema",
	Long:  `Descarga y ejecuta el script de desinstalación oficial para eliminar Kolyn y limpiar sus archivos.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUninstall(cmd.Context())
	},
}

func runUninstall(ctx context.Context) error {
	ui.ShowSection("🗑️  Kolyn Uninstall")

	ui.YellowText.Println("⚠️  ADVERTENCIA: Esto eliminará el ejecutable de Kolyn.")
	ui.Gray.Println("El script de desinstalación te preguntará si también deseas borrar")
	ui.Gray.Println("tus configuraciones y datos (skills, servicios Docker, etc).")
	fmt.Println()

	ui.WhiteText.Print("¿Estás seguro de que deseas continuar? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	answer, err := readInput(reader)
	if err != nil {
		return fmt.Errorf("error leyendo entrada: %w", err)
	}

	if strings.ToLower(answer) != "y" && strings.ToLower(answer) != "yes" && strings.ToLower(answer) != "s" && strings.ToLower(answer) != "si" {
		ui.PrintInfo("Operación cancelada.")
		return nil
	}

	// Determine script URL based on OS
	var scriptURL string
	var shell string
	var args []string
	var ext string

	if runtime.GOOS == "windows" {
		scriptURL = "https://raw.githubusercontent.com/isai-arellano/kolyn-cli/main/uninstall.ps1"
		shell = "powershell"
		ext = ".ps1"
		// We pass -NoExit to keep the window open if it were a new window,
		// but here we want it to run in current console.
		// ExecutionPolicy Bypass is needed.
		args = []string{"-ExecutionPolicy", "Bypass", "-File"}
	} else {
		scriptURL = "https://raw.githubusercontent.com/isai-arellano/kolyn-cli/main/uninstall.sh"
		shell = "/bin/sh"
		ext = ".sh"
		args = []string{}
	}

	ui.PrintStep("Descargando desinstalador...")

	// Download script
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, scriptURL, nil)
	if err != nil {
		return fmt.Errorf("error creando petición: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error descargando script: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error HTTP: %d", resp.StatusCode)
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "kolyn-uninstall-*"+ext)
	if err != nil {
		return fmt.Errorf("error creando archivo temporal: %w", err)
	}
	// We do NOT remove the temp file here because the child process needs it after we exit.
	// The uninstall script typically doesn't delete itself, so it might remain in temp.
	// That's acceptable for an uninstaller.
	// Ideally the script would delete itself, but we can't easily orchestrate that
	// from the dead parent process.

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		return fmt.Errorf("error escribiendo script: %w", err)
	}
	tmpFile.Close()

	if runtime.GOOS != "windows" {
		os.Chmod(tmpFile.Name(), 0755)
	}

	ui.PrintStep("Iniciando desinstalación...")

	// Prepare command
	finalArgs := append(args, tmpFile.Name())
	cmd := exec.Command(shell, finalArgs...)

	// Connect streams so user can interact
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start detached
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error iniciando desinstalador: %w", err)
	}

	ui.PrintSuccess("Desinstalador iniciado.")
	ui.Gray.Println("Kolyn se cerrará ahora para permitir su eliminación.")

	// Release the file lock by exiting
	os.Exit(0)

	return nil
}
