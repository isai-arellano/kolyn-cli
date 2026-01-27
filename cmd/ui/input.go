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

	// Bilingual prompt
	prompt := fmt.Sprintf("%s [y(si)/n(no)]: ", question)
	Magenta.Print("❓ " + prompt)

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	// Accept English and Spanish affirmatives
	return input == "y" || input == "yes" || input == "s" || input == "si"
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
