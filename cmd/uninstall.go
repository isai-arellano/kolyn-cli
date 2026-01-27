package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
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
	ui.ShowSection(ui.GetText("uninstall_title"))

	ui.YellowText.Println(ui.GetText("uninstall_warning"))
	ui.Gray.Println(ui.GetText("uninstall_details"))
	fmt.Println()

	if !ui.AskYesNo(ui.GetText("uninstall_confirm")) {
		ui.PrintInfo(ui.GetText("uninstall_cancel"))
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

	ui.PrintStep(ui.GetText("uninstall_downloading"))

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

	ui.PrintStep(ui.GetText("uninstall_starting"))

	// Prepare command
	finalArgs := append(args, tmpFile.Name())

	ui.PrintSuccess(ui.GetText("uninstall_started"))
	ui.Gray.Println(ui.GetText("uninstall_closing"))

	// Replace process to preserve TTY for sudo
	if err := replaceProcess(shell, finalArgs); err != nil {
		return fmt.Errorf("error reemplazando proceso: %w", err)
	}

	return nil
}
