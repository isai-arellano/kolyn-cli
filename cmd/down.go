package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/isai-arellano/kolyn-cli/cmd/ui"
	"github.com/spf13/cobra"
)

var dockerDownCmd = &cobra.Command{
	Use:     "down",
	Short:   "Detiene servicios Docker",
	Long:    `Detiene y elimina los contenedores de servicios Docker levantados con kolyn docker up.`,
	Aliases: []string{"docker-down"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDockerDownCommand(cmd.Context())
	},
}

var statusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Muestra el estado de los servicios Docker",
	Long:    `Lista todos los servicios Docker configurados y muestra si están corriendo o detenidos.`,
	Aliases: []string{"list", "ls", "docker-list"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStatusCommand(cmd.Context())
	},
}

type ServiceInfo struct {
	Name string
	Path string
}

func runStatusCommand(ctx context.Context) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error obteniendo home dir: %w", err)
	}
	dockerDir := filepath.Join(homeDir, ".kolyn", "services")

	services, err := getExistingServices(dockerDir)
	if err != nil {
		return fmt.Errorf("error escaneando servicios: %w", err)
	}

	ui.ShowSection("📋 Servicios Docker")

	if len(services) == 0 {
		ui.Gray.Println("  No hay servicios configurados en ~/.kolyn/services/")
		ui.Gray.Println("  Ejecuta 'kolyn up' para crear uno.")
		return nil
	}

	runningCount := 0

	fmt.Println()
	fmt.Printf("  %-35s %s\n", "SERVICIO", "ESTADO")
	ui.Gray.Println("  " + strings.Repeat("─", 50))

	for _, s := range services {
		status := getServiceStatus(ctx, s.Path)
		if status == "running" {
			ui.Green.Printf("  %-35s ● running\n", s.Name)
			runningCount++
		} else if status == "stopped" {
			ui.Gray.Printf("  %-35s ○ stopped\n", s.Name)
		} else {
			ui.Yellow.Printf("  %-35s ? unknown\n", s.Name)
		}
	}

	fmt.Println()
	ui.Gray.Printf("  Total: %d servicio(s), %d ejecutándose\n", len(services), runningCount)

	return nil
}

func runDockerDownCommand(ctx context.Context) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error obteniendo home dir: %w", err)
	}
	dockerDir := filepath.Join(homeDir, ".kolyn", "services")

	services, err := getExistingServices(dockerDir)
	if err != nil {
		return fmt.Errorf("error escaneando servicios: %w", err)
	}

	if len(services) == 0 {
		ui.PrintInfo("No hay servicios levantados en ~/.kolyn/services/")
		return nil
	}

	ui.ShowSection("🛑 Kolyn Down - Detener Servicios")

	fmt.Println("Servicios disponibles:\n")

	for i, s := range services {
		status := getServiceStatus(ctx, s.Path)
		if status == "running" {
			ui.Green.Printf("  %d. %-30s [RUNNING]\n", i+1, s.Name)
		} else {
			ui.Gray.Printf("  %d. %-30s [%s]\n", i+1, s.Name, status)
		}
	}
	ui.Gray.Println("  0. Cancelar")
	fmt.Println()

	fmt.Print("Selecciona servicio a detener: ")

	reader := bufio.NewReader(os.Stdin)
	selection, err := readInput(reader)
	if err != nil {
		return err
	}

	if selection == "0" || selection == "" {
		ui.PrintInfo("Operación cancelada")
		return nil
	}

	var idx int
	if _, err := fmt.Sscan(selection, &idx); err != nil {
		ui.PrintWarning("Selección inválida")
		return nil
	}

	if idx < 1 || idx > len(services) {
		ui.PrintWarning("Selección inválida")
		return nil
	}

	service := services[idx-1]
	return stopService(ctx, service)
}

func getExistingServices(basePath string) ([]ServiceInfo, error) {
	var services []ServiceInfo

	entries, err := os.ReadDir(basePath)
	if err != nil {
		// If dir doesn't exist, just return empty list
		if os.IsNotExist(err) {
			return services, nil
		}
		return services, nil
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		servicePath := filepath.Join(basePath, entry.Name())
		composePath := filepath.Join(servicePath, "docker-compose.yml")

		if _, err := os.Stat(composePath); err == nil {
			cleanName := entry.Name()
			cleanName = strings.ReplaceAll(cleanName, "-", " ")
			// Simple Title Case logic
			if len(cleanName) > 0 {
				cleanName = strings.ToUpper(cleanName[:1]) + cleanName[1:]
			}

			services = append(services, ServiceInfo{
				Name: cleanName,
				Path: servicePath,
			})
		}
	}

	return services, nil
}

func getServiceStatus(ctx context.Context, servicePath string) string {
	cmd := exec.CommandContext(ctx, "docker", "compose", "ps", "-q")
	cmd.Dir = servicePath
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	containerID := strings.TrimSpace(string(output))
	if containerID == "" {
		return "stopped"
	}

	checkCmd := exec.CommandContext(ctx, "docker", "inspect", "-f", "{{.State.Running}}", containerID)
	checkCmd.Dir = servicePath
	statusOutput, err := checkCmd.Output()
	if err != nil {
		return "unknown"
	}

	if strings.Contains(string(statusOutput), "true") {
		return "running"
	}
	return "stopped"
}

func stopService(ctx context.Context, s ServiceInfo) error {
	composePath := filepath.Join(s.Path, "docker-compose.yml")
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		ui.PrintWarning("No se encontró docker-compose.yml")
		return nil
	}

	ui.PrintStep("Deteniendo servicio '%s'...", s.Name)

	cmd := exec.CommandContext(ctx, "docker", "compose", "down", "-v")
	cmd.Dir = s.Path
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error ejecutando docker compose down: %w", err)
	}

	ui.PrintSuccess("Servicio '%s' detenido!", s.Name)
	ui.Gray.Printf("  Directorio: %s\n", s.Path)
	ui.Gray.Println("  (El compose y archivos se mantienen)")

	return nil
}
