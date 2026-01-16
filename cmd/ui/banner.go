package ui

import (
	"fmt"
)

const version = "v0.1.0"

func ShowBanner() {
	banner := `
╔══════════════════════════════════════════════════════════════════╗
║                                                                  ║
║   ██╗  ██╗ ██████╗ ██╗  ██╗   ██╗███╗   ██╗                    ║
║   ██║ ██╔╝██╔═══██╗██║  ╚██╗ ██╔╝████╗  ██║                    ║
║   █████╔╝ ██║   ██║██║   ╚████╔╝ ██╔██╗ ██║                    ║
║   ██╔═██╗ ██║   ██║██║    ╚██╔╝  ██║╚██╗██║                    ║
║   ██║  ██╗╚██████╔╝███████╗██║   ██║ ╚████║                    ║
║   ╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚═╝   ╚═╝  ╚═══╝                    ║
║                                                                  ║
║           Orquestador de Desarrollo con IA                      ║
║                                                                  ║
╚══════════════════════════════════════════════════════════════════╝
`
	Cyan.Println(banner)
	Gray.Printf("                                    %s\n\n", version)
}

func ShowSection(title string) {
	fmt.Println()
	Cyan.Println(title)
	Separator()
}
