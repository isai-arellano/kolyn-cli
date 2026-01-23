package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/isai-arellano/kolyn-cli/cmd/ui"
	"github.com/spf13/cobra"
)

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Gestiona llaves y configuraciones SSH",
	Long:  `Herramientas para generar llaves SSH y configurar accesos a servidores de forma estandarizada.`,
}

var sshCreateCmd = &cobra.Command{
	Use:   "create [nombre] [ip] [usuario]",
	Short: "Crea una nueva llave SSH y configura el acceso",
	Args:  cobra.MinimumNArgs(2),
	Long: `Genera una nueva llave SSH (ed25519), la guarda en ~/.ssh/ y actualiza ~/.ssh/config.
	
Ejemplo:
  kolyn tools ssh create mi-cliente 192.168.1.50 root`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		ip := args[1]
		user := "root"
		if len(args) > 2 {
			user = args[2]
		}
		return runSshCreate(name, ip, user)
	},
}

func runSshCreate(name, ip, user string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error obteniendo home: %w", err)
	}

	sshDir := filepath.Join(homeDir, ".ssh")
	keyPath := filepath.Join(sshDir, name)
	configPath := filepath.Join(sshDir, "config")

	ui.ShowSection(fmt.Sprintf("🔑 Generando acceso SSH: %s", name))

	// 1. Verificar si la llave ya existe
	if _, err := os.Stat(keyPath); err == nil {
		ui.PrintWarning("La llave '%s' ya existe en ~/.ssh/", name)
		return fmt.Errorf("operación abortada para no sobrescribir llaves existentes")
	}

	// 2. Generar llave ED25519
	ui.PrintStep("Generando par de llaves ED25519...")
	// ssh-keygen -t ed25519 -f keyPath -C "generada por kolyn para name" -N ""
	cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-f", keyPath, "-C", fmt.Sprintf("kolyn-%s", name), "-N", "")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error generando llave: %s", string(output))
	}
	ui.PrintSuccess("Llaves generadas en: %s", keyPath)

	// 3. Actualizar ~/.ssh/config
	ui.PrintStep("Actualizando ~/.ssh/config...")

	configEntry := fmt.Sprintf(`
# Generado por Kolyn (%s)
Host %s
  HostName %s
  User %s
  IdentityFile %s
  IdentitiesOnly yes
`, name, name, ip, user, keyPath)

	f, err := os.OpenFile(configPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error abriendo config ssh: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(configEntry); err != nil {
		return fmt.Errorf("error escribiendo en config: %w", err)
	}
	ui.PrintSuccess("Configuración actualizada")

	// 4. Copiar llave al servidor (Opcional)
	fmt.Println()
	ui.YellowText.Printf("¿Deseas copiar la llave pública al servidor %s? [y/N]: ", ip)
	fmt.Print("> ")

	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" || strings.ToLower(response) == "s" || strings.ToLower(response) == "si" {
		ui.PrintStep("Copiando llave pública (te pedirá la contraseña del servidor)...")

		// ssh-copy-id -i keyPath.pub user@ip
		copyCmd := exec.Command("ssh-copy-id", "-i", keyPath+".pub", fmt.Sprintf("%s@%s", user, ip))
		copyCmd.Stdin = os.Stdin
		copyCmd.Stdout = os.Stdout
		copyCmd.Stderr = os.Stderr

		if err := copyCmd.Run(); err != nil {
			ui.PrintWarning("No se pudo copiar la llave automáticamente: %v", err)
			ui.Gray.Println("Puedes hacerlo manualmente con:")
			ui.Gray.Printf("ssh-copy-id -i %s.pub %s@%s\n", keyPath, user, ip)
		} else {
			ui.PrintSuccess("¡Acceso configurado exitosamente!")
			ui.Gray.Printf("\nPrueba conectar con: ssh %s\n", name)
		}
	} else {
		ui.PrintInfo("Paso omitido. Recuerda agregar la llave pública manualmente.")
		ui.Gray.Printf("Contenido de %s.pub:\n", keyPath)
		pubContent, _ := os.ReadFile(keyPath + ".pub")
		fmt.Println(string(pubContent))
	}

	return nil
}
