package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourusername/kolyn/cmd/ui"
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Levanta servicios Docker desde templates",
	Long:  `Crea y levanta servicios Docker desde templates pre-configurados.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUpCommand()
	},
}

type ComposeTemplate struct {
	Name        string
	Description string
	Service     string
	Port        string
}

func getTemplatesFromSkills() []ComposeTemplate {
	homeDir, _ := os.UserHomeDir()
	skillsPath := filepath.Join(homeDir, ".kolyn/skills", "automation", "docker-compose-automation.md")

	content, err := os.ReadFile(skillsPath)
	if err != nil {
		return getDefaultTemplates()
	}

	var templates []ComposeTemplate

	sections := strings.Split(string(content), "### ")
	for _, section := range sections[1:] {
		lines := strings.Split(section, "\n")
		if len(lines) < 2 {
			continue
		}

		name := strings.TrimSpace(lines[0])
		name = strings.TrimPrefix(name, "1. ")
		name = strings.TrimPrefix(name, "2. ")
		name = strings.TrimPrefix(name, "3. ")
		name = strings.TrimPrefix(name, "4. ")
		name = strings.TrimPrefix(name, "5. ")
		name = strings.TrimPrefix(name, "6. ")
		name = strings.TrimPrefix(name, "7. ")
		name = strings.TrimPrefix(name, "8. ")
		name = strings.TrimPrefix(name, "9. ")
		name = strings.TrimSpace(name)
		sectionContent := strings.Join(lines[1:], "\n")

		if name == "" || !strings.HasPrefix(sectionContent, "```yaml") {
			continue
		}

		var port string
		if strings.Contains(name, "n8n") {
			port = "5678"
		} else if strings.Contains(name, "PostgreSQL") && strings.Contains(name, "pgAdmin") {
			port = "5050"
		} else if strings.Contains(name, "PostgreSQL") {
			port = "5432"
		} else if strings.Contains(name, "Redis") {
			port = "6379"
		} else if strings.Contains(name, "MongoDB") {
			port = "27017"
		} else if strings.Contains(name, "Next.js") {
			port = "3000"
		} else {
			port = "?"
		}

		serviceName := strings.ToLower(name)
		serviceName = strings.ReplaceAll(serviceName, " ", "-")
		serviceName = strings.ReplaceAll(serviceName, "+", "-")

		templates = append(templates, ComposeTemplate{
			Name:        name,
			Description: name,
			Service:     serviceName,
			Port:        port,
		})
	}

	if len(templates) == 0 {
		return getDefaultTemplates()
	}

	return templates
}

func getDefaultTemplates() []ComposeTemplate {
	return []ComposeTemplate{
		{"n8n", "Automatización de workflows", "n8n", "5678"},
		{"postgres", "Base de datos PostgreSQL", "postgres", "5432"},
		{"postgres-pgadmin", "PostgreSQL + pgAdmin", "postgres-pgadmin", "5050"},
		{"redis", "Cache en memoria", "redis", "6379"},
		{"mongodb", "Base de datos MongoDB", "mongodb", "27017"},
		{"nextjs-postgres", "Next.js + PostgreSQL + Redis", "nextjs-postgres", "3000"},
	}
}

func runUpCommand() error {
	templates := getTemplatesFromSkills()

	ui.ShowSection("🚀 Kolyn Up - Levantar Servicios")

	fmt.Println("Selecciona un servicio para levantar:\n")

	for i, t := range templates {
		ui.WhiteText.Printf("  %d. %-25s (puerto: %s)\n", i+1, t.Name, t.Port)
	}
	ui.Gray.Println("  0. Cancelar")
	fmt.Println()

	fmt.Print("Selecciona: ")
	selection := readInput()

	if selection == "0" || selection == "" {
		ui.PrintInfo("Operación cancelada")
		return nil
	}

	var idx int
	fmt.Sscan(selection, &idx)
	if idx < 1 || idx > len(templates) {
		ui.PrintWarning("Selección inválida")
		return nil
	}

	template := templates[idx-1]
	return startService(template, templates)
}

func startService(t ComposeTemplate, allTemplates []ComposeTemplate) error {
	homeDir, _ := os.UserHomeDir()
	dockerDir := filepath.Join(homeDir, "docker", t.Service)

	if _, err := os.Stat(dockerDir); err == nil {
		ui.PrintWarning("El servicio '%s' ya existe en: %s", t.Name, dockerDir)
		fmt.Print("¿Deseas sobreescribir? [s/n]: ")
		answer := readInput()
		if strings.ToLower(answer) != "s" && strings.ToLower(answer) != "si" && strings.ToLower(answer) != "yes" {
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

	content := generateCompose(t.Service, allTemplates)
	if err := os.WriteFile(composePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("error escribiendo compose: %w", err)
	}

	ui.PrintSuccess("docker-compose.yml creado!")

	fmt.Println()
	ui.YellowText.Println("¿Deseas levantar el servicio ahora? [s/n]: ")
	fmt.Print("> ")
	answer := readInput()

	if strings.ToLower(answer) == "s" || strings.ToLower(answer) == "si" || strings.ToLower(answer) == "yes" {
		ui.PrintStep("Levantando servicio con Docker...")

		cmd := exec.Command("docker", "compose", "up", "-d")
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

func generateCompose(service string, templates []ComposeTemplate) string {
	for _, t := range templates {
		if t.Service == service {
			homeDir, _ := os.UserHomeDir()
			skillsPath := filepath.Join(homeDir, ".kolyn/skills", "automation", "docker-compose-automation.md")
			content, err := os.ReadFile(skillsPath)
			if err != nil {
				return getDefaultCompose(service)
			}

			sectionName := t.Name
			sections := strings.Split(string(content), "### ")
			for _, section := range sections {
				if strings.HasPrefix(section, sectionName) {
					lines := strings.Split(section, "\n")
					var yamlLines []string
					inYaml := false
					for _, line := range lines {
						if strings.HasPrefix(line, "```yaml") {
							inYaml = true
							continue
						}
						if strings.HasPrefix(line, "```") && inYaml {
							break
						}
						if inYaml {
							yamlLines = append(yamlLines, line)
						}
					}
					if len(yamlLines) > 0 {
						return strings.Join(yamlLines, "\n") + "\n"
					}
				}
			}
		}
	}
	return getDefaultCompose(service)
}

func getDefaultCompose(service string) string {
	switch service {
	case "n8n":
		return `version: '3.8'

services:
  n8n:
    image: n8nio/n8n
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
`
	case "postgres":
		return `version: '3.8'

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
`
	case "postgres-pgadmin":
		return `version: '3.8'

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

  pgadmin:
    image: dpage/pgadmin4:8
    container_name: pgadmin
    restart: unless-stopped
    ports:
      - "5050:80"
    environment:
      - PGADMIN_DEFAULT_EMAIL=admin@admin.com
      - PGADMIN_DEFAULT_PASSWORD=admin
    volumes:
      - pgadmin_data:/var/lib/pgadmin

volumes:
  postgres_data:
  pgadmin_data:
`
	case "redis":
		return `version: '3.8'

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
`
	case "mongodb":
		return `version: '3.8'

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
`
	case "nextjs-postgres":
		return `version: '3.8'

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: nextjs-app
    restart: unless-stopped
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=development
    volumes:
      - .:/app
      - /app/node_modules
    depends_on:
      - postgres
      - redis

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

  redis:
    image: redis:7-alpine
    container_name: redis
    restart: unless-stopped
    ports:
      - "6379:6379"

volumes:
  postgres_data:
`
	default:
		return ""
	}
}
