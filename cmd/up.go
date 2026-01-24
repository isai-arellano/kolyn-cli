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

var dockerUpCmd = &cobra.Command{
	Use:     "up",
	Short:   "Levanta servicios Docker desde templates",
	Long:    `Crea y levanta servicios Docker desde templates locales (~/.kolyn/templates).`,
	Aliases: []string{"docker-up"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDockerUpCommand(cmd.Context())
	},
}

type ComposeTemplate struct {
	Name        string
	Description string
	Service     string
	Port        string
	Content     string
}

func getTemplates() ([]ComposeTemplate, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo home dir: %w", err)
	}
	templatesDir := filepath.Join(homeDir, ".kolyn", "templates")

	// 1. Inicializar directorio y defaults si no existe
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		if err := os.MkdirAll(templatesDir, 0755); err == nil {
			if err := writeDefaultTemplates(templatesDir); err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("error creando directorio templates: %w", err)
		}
	}

	// 2. Escanear archivos .yml / .yaml
	var templates []ComposeTemplate
	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		return nil, fmt.Errorf("error leyendo directorio templates: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".yml" && ext != ".yaml" {
			continue
		}

		path := filepath.Join(templatesDir, entry.Name())
		contentBytes, err := os.ReadFile(path)
		if err != nil {
			continue // Skip unreadable
		}
		content := string(contentBytes)

		name := strings.TrimSuffix(entry.Name(), ext)
		port := extractPortFromCompose(content)

		templates = append(templates, ComposeTemplate{
			Name:        name,
			Description: fmt.Sprintf("Template: %s", entry.Name()),
			Service:     name,
			Port:        port,
			Content:     content,
		})
	}

	return templates, nil
}

func writeDefaultTemplates(dir string) error {
	defaults := map[string]string{
		"n8n.yml": `services:
  n8n:
    image: n8nio/n8n:latest
    container_name: n8n
    restart: unless-stopped
    ports:
      - "5678:5678"
    environment:
      - N8N_BASIC_AUTH_ACTIVE=true
      - N8N_BASIC_AUTH_USER=admin
      - N8N_BASIC_AUTH_PASSWORD=password123
      - WEBHOOK_URL=http://localhost:5678/
      - GENERIC_TIMEZONE=America/Argentina/Buenos_Aires
      - TZ=America/Argentina/Buenos_Aires
    volumes:
      - n8n_data:/home/node/.n8n
      - ./local_files:/files
    networks:
      - n8n-network

volumes:
  n8n_data:

networks:
  n8n-network:
    driver: bridge
`,
		"postgres.yml": `version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    container_name: postgres
    restart: unless-stopped
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=postgres
      - POSTGRES_DB=app
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
`,
		"redis.yml": `version: '3.8'

services:
  redis:
    image: redis:7-alpine
    container_name: redis
    restart: unless-stopped
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

volumes:
  redis_data:
`,
		"mongodb.yml": `version: '3.8'

services:
  mongodb:
    image: mongo:7
    container_name: mongodb
    restart: unless-stopped
    ports:
      - "27017:27017"
    environment:
      - MONGO_INITDB_DATABASE=app
      - MONGO_INITDB_ROOT_USERNAME=admin
      - MONGO_INITDB_ROOT_PASSWORD=password
    volumes:
      - mongo_data:/data/db

  mongo-express:
    image: mongo-express:1.0
    container_name: mongo-express
    restart: unless-stopped
    ports:
      - "8081:8081"
    environment:
      - ME_CONFIG_MONGODB_URL=mongodb://admin:password@mongodb:27017
    depends_on:
      - mongodb

volumes:
  mongo_data:
`,
	}

	for filename, content := range defaults {
		if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0644); err != nil {
			return fmt.Errorf("error escribiendo template default %s: %w", filename, err)
		}
	}
	return nil
}

func extractPortFromCompose(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, "-") && strings.Contains(trimmed, ":") {
			parts := strings.Split(trimmed, ":")
			if len(parts) >= 2 {
				// Cleaner extraction logic could be regex, but keeping simple
				portCandidate := strings.Trim(parts[0], " -\"'")
				if len(portCandidate) > 0 && portCandidate[0] >= '0' && portCandidate[0] <= '9' {
					return portCandidate
				}
			}
		}
	}
	return "?"
}

func runDockerUpCommand(ctx context.Context) error {
	templates, err := getTemplates()
	if err != nil {
		return err
	}

	ui.ShowSection("🚀 Kolyn Up - Levantar Servicios")

	if len(templates) == 0 {
		ui.PrintWarning("No se encontraron templates en ~/.kolyn/templates/")
		return nil
	}

	fmt.Println("Selecciona un servicio para levantar:\n")

	for i, t := range templates {
		ui.WhiteText.Printf("  %d. %-25s (puerto: %s)\n", i+1, t.Name, t.Port)
	}
	ui.Gray.Println("  0. Cancelar")
	fmt.Println()

	homeDir, _ := os.UserHomeDir()
	ui.Gray.Printf("💡 Tip: Agrega tus propios .yml en %s\n\n", filepath.Join(homeDir, ".kolyn", "templates"))

	fmt.Print("Selecciona: ")

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

	if idx < 1 || idx > len(templates) {
		ui.PrintWarning("Selección inválida")
		return nil
	}

	template := templates[idx-1]
	return startService(ctx, template, reader)
}

func startService(ctx context.Context, t ComposeTemplate, reader *bufio.Reader) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error obteniendo home dir: %w", err)
	}
	dockerDir := filepath.Join(homeDir, ".kolyn", "services", t.Service)

	if _, err := os.Stat(dockerDir); err == nil {
		ui.PrintWarning("El servicio '%s' ya existe en: %s", t.Name, dockerDir)
		fmt.Println()
		ui.WhiteText.Println("  1. Sobrescribir (regenerar compose)")
		ui.WhiteText.Println("  2. Levantar (iniciar servicio existente)")
		ui.Gray.Println("  0. Cancelar")
		fmt.Println()
		fmt.Print("Selecciona una opción: ")

		answer, err := readInput(reader)
		if err != nil {
			return err
		}

		switch answer {
		case "1":
			ui.PrintStep("Sobrescribiendo compose...")
		case "2":
			return liftExistingService(ctx, dockerDir, t)
		default:
			ui.PrintInfo("Operación cancelada")
			return nil
		}
	} else {
		ui.PrintStep("Creando directorio: %s", dockerDir)
		if err := os.MkdirAll(dockerDir, 0755); err != nil {
			return fmt.Errorf("error creando directorio: %w", err)
		}
	}

	composePath := filepath.Join(dockerDir, "docker-compose.yml")
	ui.PrintStep("Generando %s...", composePath)

	// Usar el contenido directo del template
	if err := os.WriteFile(composePath, []byte(t.Content), 0644); err != nil {
		return fmt.Errorf("error escribiendo compose: %w", err)
	}

	ui.PrintSuccess("docker-compose.yml creado!")

	fmt.Println()
	ui.YellowText.Println("¿Deseas levantar el servicio ahora? [s/n]: ")
	fmt.Print("> ")

	answer, err := readInput(reader)
	if err != nil {
		return err
	}

	if strings.ToLower(answer) == "s" || strings.ToLower(answer) == "si" || strings.ToLower(answer) == "yes" {
		ui.PrintStep("Levantando servicio con Docker...")

		cmd := exec.CommandContext(ctx, "docker", "compose", "up", "-d")
		cmd.Dir = dockerDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error ejecutando docker compose: %w", err)
		}

		ui.PrintSuccess("Servicio '%s' levantado!", t.Name)
		ui.Gray.Println("\nComandos útiles:")
		ui.Gray.Printf("  Ver logs:    cd %s && docker compose logs -f\n", dockerDir)
		ui.Gray.Printf("  Detener:     cd %s && docker compose down\n", dockerDir)
		ui.Gray.Printf("  Acceder:     http://localhost:%s\n", t.Port)
	}

	return nil
}

func liftExistingService(ctx context.Context, dockerDir string, t ComposeTemplate) error {
	composePath := filepath.Join(dockerDir, "docker-compose.yml")
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		ui.PrintWarning("No se encontró docker-compose.yml en: %s", dockerDir)
		return nil
	}

	ui.PrintStep("Levantando servicio existente...")

	cmd := exec.CommandContext(ctx, "docker", "compose", "up", "-d")
	cmd.Dir = dockerDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error ejecutando docker compose: %w", err)
	}

	ui.PrintSuccess("Servicio '%s' levantado!", t.Name)
	ui.Gray.Println("\nComandos útiles:")
	ui.Gray.Printf("  Ver logs:    cd %s && docker compose logs -f\n", dockerDir)
	ui.Gray.Printf("  Detener:     cd %s && docker compose down\n", dockerDir)
	ui.Gray.Printf("  Acceder:     http://localhost:%s\n", t.Port)

	return nil
}
