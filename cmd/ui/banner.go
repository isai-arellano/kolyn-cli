package ui

import (
	"fmt"
)

func ShowBanner(version string) {
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
