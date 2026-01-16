package ui

import "github.com/fatih/color"

// Paleta de colores de Kolyn CLI
var (
	// Títulos y encabezados
	Cyan     = color.New(color.FgCyan, color.Bold)
	CyanText = color.New(color.FgCyan)

	// Éxito y confirmaciones
	Green     = color.New(color.FgGreen, color.Bold)
	GreenText = color.New(color.FgGreen)

	// Advertencias y opciones
	Yellow     = color.New(color.FgYellow, color.Bold)
	YellowText = color.New(color.FgYellow)

	// Información y proceso
	Blue     = color.New(color.FgBlue, color.Bold)
	BlueText = color.New(color.FgBlue)

	// Prompts y preguntas
	Magenta     = color.New(color.FgMagenta, color.Bold)
	MagentaText = color.New(color.FgMagenta)

	// Errores
	Red     = color.New(color.FgRed, color.Bold)
	RedText = color.New(color.FgRed)

	// Énfasis
	White     = color.New(color.FgWhite, color.Bold)
	WhiteText = color.New(color.FgWhite)

	// Gris para texto secundario
	Gray = color.New(color.FgHiBlack)

	// Combinaciones especiales
	Success = color.New(color.FgGreen, color.Bold)
	Error   = color.New(color.FgRed, color.Bold)
	Warning = color.New(color.FgYellow, color.Bold)
	Info    = color.New(color.FgCyan)
)

// Funciones helper para imprimir con colores

// PrintSuccess imprime un mensaje de éxito
func PrintSuccess(msg string, args ...interface{}) {
	Success.Printf("✅ "+msg+"\n", args...)
}

// PrintError imprime un mensaje de error
func PrintError(msg string, args ...interface{}) {
	Error.Printf("❌ "+msg+"\n", args...)
}

// PrintWarning imprime una advertencia
func PrintWarning(msg string, args ...interface{}) {
	Warning.Printf("⚠️  "+msg+"\n", args...)
}

// PrintInfo imprime información
func PrintInfo(msg string, args ...interface{}) {
	Info.Printf("ℹ️  "+msg+"\n", args...)
}

// PrintStep imprime un paso del proceso
func PrintStep(msg string, args ...interface{}) {
	Cyan.Printf("▸ "+msg+"\n", args...)
}

// PrintQuestion imprime una pregunta
func PrintQuestion(msg string, args ...interface{}) {
	Magenta.Printf("❓ "+msg+"\n", args...)
}

// Separator imprime una línea separadora
func Separator() {
	Gray.Println("──────────────────────────────────────────────────────────────────")
}

// SeparatorDouble imprime una línea separadora doble
func SeparatorDouble() {
	Gray.Println("══════════════════════════════════════════════════════════════════")
}
