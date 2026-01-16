package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourusername/kolyn/cmd/ui"
)

var dockerDownCmd = &cobra.Command{
	Use:     "down",
	Short:   "Detiene servicios Docker",
	Long:    `Detiene y elimina los contenedores de servicios Docker levantados con kolyn docker up.`,
	Aliases: []string{"docker-down"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDockerDownCommand()
	},
}

type ServiceInfo struct {
	Name string
	Path string
}

func readInputFromStdin() string {
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func runDockerDownCommand() error {
	homeDir, _ := os.UserHomeDir()
	dockerDir := filepath.Join(homeDir, "docker")

	services, err := getExistingServices(dockerDir)
	if err != nil {
		return fmt.Errorf("error escaneando servicios: %w", err)
	}

	if len(services) == 0 {
		ui.PrintInfo("No hay servicios levantados en ~/docker/")
		return nil
	}

	ui.ShowSection("🛑 Kolyn Down - Detener Servicios")

	fmt.Println("Servicios disponibles:\n")

	for i, s := range services {
		status := getServiceStatus(s.Path)
		if status == "running" {
			ui.Green.Printf("  %d. %-30s [RUNNING]\n", i+1, s.Name)
		} else {
			ui.Gray.Printf("  %d. %-30s [%s]\n", i+1, s.Name, status)
		}
	}
	ui.Gray.Println("  0. Cancelar")
	fmt.Println()

	fmt.Print("Selecciona servicio a detener: ")
	selection := readInputFromStdin()

	if selection == "0" || selection == "" {
		ui.PrintInfo("Operación cancelada")
		return nil
	}

	var idx int
	fmt.Sscan(selection, &idx)
	if idx < 1 || idx > len(services) {
		ui.PrintWarning("Selección inválida")
		return nil
	}

	service := services[idx-1]
	return stopService(service)
}

func getExistingServices(basePath string) ([]ServiceInfo, error) {
	var services []ServiceInfo

	entries, err := os.ReadDir(basePath)
	if err != nil {
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
			cleanName = strings.Title(cleanName)
			services = append(services, ServiceInfo{
				Name: cleanName,
				Path: servicePath,
			})
		}
	}

	return services, nil
}

func getServiceStatus(servicePath string) string {
	cmd := exec.Command("docker", "compose", "ps", "-q")
	cmd.Dir = servicePath
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	containerID := strings.TrimSpace(string(output))
	if containerID == "" {
		return "stopped"
	}

	checkCmd := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", containerID)
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

func stopService(s ServiceInfo) error {
	composePath := filepath.Join(s.Path, "docker-compose.yml")
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		ui.PrintWarning("No se encontró docker-compose.yml")
		return nil
	}

	ui.PrintStep("Deteniendo servicio '%s'...", s.Name)

	cmd := exec.Command("docker", "compose", "down", "-v")
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
