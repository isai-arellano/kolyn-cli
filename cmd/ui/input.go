package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var CurrentLanguage = "es"

// AskYesNo prompts the user with a bilingual yes/no question.
// Returns true for yes, false for no.
func AskYesNo(question string) bool {
	reader := bufio.NewReader(os.Stdin)

	// Dynamic prompt based on language
	var suffix string
	if CurrentLanguage == "en" {
		suffix = "[y/n]"
	} else {
		suffix = "[s/n]"
	}

	prompt := fmt.Sprintf("%s %s: ", question, suffix)
	Magenta.Print("❓ " + prompt)

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if CurrentLanguage == "en" {
		return input == "y" || input == "yes"
	}
	// Spanish logic
	return input == "s" || input == "si"
}

// ReadInput reads a line of text from the user
func ReadInput(prompt string) string {
	if prompt != "" {
		fmt.Print(prompt)
	}
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// SelectLanguage prompts the user to choose a language.
func SelectLanguage() string {
	Cyan.Println("\n🌐 Select Language / Selecciona Idioma:")
	fmt.Println("  1. English")
	fmt.Println("  2. Español (México)")

	for {
		choice := ReadInput("> ")
		switch choice {
		case "1", "en", "english":
			return "en"
		case "2", "es", "español", "spanish":
			return "es"
		default:
			PrintWarning("Invalid option / Opción inválida")
		}
	}
}

// GetText returns translated text based on CurrentLanguage
func GetText(key string, args ...interface{}) string {
	es := map[string]string{
		"sync_start":        "Iniciando sincronización...",
		"sync_success":      "Sincronización completada exitosamente.",
		"no_config":         "No se detectó configuración. Iniciando configuración global.",
		"global_created":    "Configuración global creada en ~/.kolyn/config.json",
		"using_global":      "Usando configuración global de skills.",
		"using_local":       "Usando configuración local del proyecto (.kolyn.json).",
		"installing_skills": "Instalando skills desde: %s",
		"updating_skills":   "Actualizando skills en: %s",
		"repo_access_error": "Error de acceso al repositorio. Si es privado, verifica tus llaves SSH o credenciales.",
		"check_start":       "🕵️  Kolyn Check - Auditoría de Proyecto",
		"no_package_json":   "No se encontró package.json. Se omitirán chequeos de dependencias.",
		"no_skills":         "No hay skills instaladas para auditar.",
		"evaluating_skill":  "\nEvaluando Skill: %s/%s",
		"missing_dep":       "  ❌ Falta dependencia: %s",
		"found_dep":         "  ✅ Dependencia encontrada: %s",
		"forbidden_dep":     "  ❌ Dependencia prohibida detectada: %s",
		"missing_file":      "  ❌ Falta archivo: %s",
		"found_file":        "  ✅ Archivo encontrado: %s",
		"audit_summary":     "Resumen: %d verificaciones, %d pasadas, %d alertas",
		"audit_issues":      "se encontraron %d problemas en la auditoría",

		// Config
		"skills_repo_prompt": "Ingresa la URL del repositorio de skills de tu equipo (ej. git@github.com:org/skills.git):",
		"using_default_repo": "Usando repositorio oficial de Kolyn.",

		// Uninstall
		"uninstall_title":       "🗑️  Kolyn Uninstall",
		"uninstall_warning":     "⚠️  ADVERTENCIA: Esto eliminará el ejecutable de Kolyn.",
		"uninstall_details":     "El script de desinstalación te preguntará si también deseas borrar\ntus configuraciones y datos (skills, servicios Docker, etc).",
		"uninstall_confirm":     "¿Estás seguro de que deseas continuar?",
		"uninstall_cancel":      "Operación cancelada.",
		"uninstall_downloading": "Descargando desinstalador...",
		"uninstall_starting":    "Iniciando desinstalación...",
		"uninstall_started":     "Desinstalador iniciado.",
		"uninstall_closing":     "Kolyn se cerrará ahora para permitir su eliminación.",

		// Docker Up
		"docker_up_title":         "🚀 Kolyn Up - Levantar Servicios",
		"docker_up_no_templates":  "No se encontraron templates en ~/.kolyn/templates/",
		"docker_up_select":        "Selecciona un servicio para levantar:\n",
		"docker_up_port":          "(puerto: %s)",
		"docker_up_cancel_opt":    "  0. Cancelar",
		"docker_up_tip":           "💡 Tip: Agrega tus propios .yml en %s\n",
		"docker_up_input":         "Selecciona: ",
		"docker_up_invalid":       "Selección inválida",
		"docker_up_exists":        "El servicio '%s' ya existe en: %s",
		"docker_up_overwrite":     "  1. Sobrescribir (regenerar compose)",
		"docker_up_lift":          "  2. Levantar (iniciar servicio existente)",
		"docker_up_overwriting":   "Sobrescribiendo compose...",
		"docker_up_generating":    "Generando %s...",
		"docker_up_created":       "docker-compose.yml creado!",
		"docker_up_confirm_start": "¿Deseas levantar el servicio ahora?",
		"docker_up_starting":      "Levantando servicio con Docker...",
		"docker_up_success":       "Servicio '%s' levantado!",
		"docker_up_cmds":          "\nComandos útiles:",
		"docker_up_logs":          "  Ver logs:    cd %s && docker compose logs -f",
		"docker_up_stop":          "  Detener:     cd %s && docker compose down",
		"docker_up_access":        "  Acceder:     http://localhost:%s",
		"docker_up_lift_existing": "Levantando servicio existente...",
	}

	en := map[string]string{
		"sync_start":        "Starting synchronization...",
		"sync_success":      "Synchronization completed successfully.",
		"no_config":         "No configuration detected. Starting global setup.",
		"global_created":    "Global configuration created at ~/.kolyn/config.json",
		"using_global":      "Using global skills configuration.",
		"using_local":       "Using local project configuration (.kolyn.json).",
		"installing_skills": "Installing skills from: %s",
		"updating_skills":   "Updating skills at: %s",
		"repo_access_error": "Repository access error. If private, check your SSH keys or credentials.",
		"check_start":       "🕵️  Kolyn Check - Project Audit",
		"no_package_json":   "package.json not found. Dependency checks skipped.",
		"no_skills":         "No installed skills found to audit.",
		"evaluating_skill":  "\nEvaluating Skill: %s/%s",
		"missing_dep":       "  ❌ Missing dependency: %s",
		"found_dep":         "  ✅ Dependency found: %s",
		"forbidden_dep":     "  ❌ Forbidden dependency detected: %s",
		"missing_file":      "  ❌ Missing file: %s",
		"found_file":        "  ✅ File found: %s",
		"audit_summary":     "Summary: %d checks, %d passed, %d warnings",
		"audit_issues":      "%d issues found during audit",

		// Config
		"skills_repo_prompt": "Enter your team's skills repository URL (e.g. git@github.com:org/skills.git):",
		"using_default_repo": "Using official Kolyn repository.",

		// Uninstall
		"uninstall_title":       "🗑️  Kolyn Uninstall",
		"uninstall_warning":     "⚠️  WARNING: This will remove the Kolyn executable.",
		"uninstall_details":     "The uninstall script will ask if you also want to remove\nyour configurations and data (skills, Docker services, etc).",
		"uninstall_confirm":     "Are you sure you want to continue?",
		"uninstall_cancel":      "Operation canceled.",
		"uninstall_downloading": "Downloading uninstaller...",
		"uninstall_starting":    "Starting uninstaller...",
		"uninstall_started":     "Uninstaller started.",
		"uninstall_closing":     "Kolyn will close now to allow removal.",

		// Docker Up
		"docker_up_title":         "🚀 Kolyn Up - Lift Services",
		"docker_up_no_templates":  "No templates found in ~/.kolyn/templates/",
		"docker_up_select":        "Select a service to lift:\n",
		"docker_up_port":          "(port: %s)",
		"docker_up_cancel_opt":    "  0. Cancel",
		"docker_up_tip":           "💡 Tip: Add your own .yml files in %s\n",
		"docker_up_input":         "Select: ",
		"docker_up_invalid":       "Invalid selection",
		"docker_up_exists":        "Service '%s' already exists at: %s",
		"docker_up_overwrite":     "  1. Overwrite (regenerate compose)",
		"docker_up_lift":          "  2. Lift (start existing service)",
		"docker_up_overwriting":   "Overwriting compose...",
		"docker_up_generating":    "Generating %s...",
		"docker_up_created":       "docker-compose.yml created!",
		"docker_up_confirm_start": "Do you want to start the service now?",
		"docker_up_starting":      "Starting service with Docker...",
		"docker_up_success":       "Service '%s' started!",
		"docker_up_cmds":          "\nUseful commands:",
		"docker_up_logs":          "  View logs:   cd %s && docker compose logs -f",
		"docker_up_stop":          "  Stop:        cd %s && docker compose down",
		"docker_up_access":        "  Access:      http://localhost:%s",
		"docker_up_lift_existing": "Starting existing service...",
	}

	var dict map[string]string
	if CurrentLanguage == "en" {
		dict = en
	} else {
		dict = es
	}

	msg, ok := dict[key]
	if !ok {
		return key // Fallback
	}

	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}
	return msg
}

// SelectIndices prompts the user to select multiple options from a list.
// Returns a slice of selected indices (0-based).
// Input format: "1, 3, 5" or "1 3 5" or "1-3" (ranges not strictly required but nice, let's stick to comma/space first).
func SelectIndices(prompt string, maxOptions int) ([]int, error) {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return []int{}, nil
	}

	// Normalize separators: replace commas with spaces
	input = strings.ReplaceAll(input, ",", " ")
	parts := strings.Fields(input)

	var selection []int
	seen := make(map[int]bool)

	for _, part := range parts {
		var idx int
		_, err := fmt.Sscan(part, &idx)
		if err != nil {
			Gray.Printf("   ⚠️  Ignorando entrada inválida: '%s'\n", part)
			continue
		}

		if idx < 1 || idx > maxOptions {
			Gray.Printf("   ⚠️  Ignorando número fuera de rango: %d (max: %d)\n", idx, maxOptions)
			continue
		}

		// Convert to 0-based index
		zeroIdx := idx - 1
		if !seen[zeroIdx] {
			selection = append(selection, zeroIdx)
			seen[zeroIdx] = true
		}
	}

	return selection, nil
}
