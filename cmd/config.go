package cmd

import (
	"context"
	"fmt"

	"github.com/isai-arellano/kolyn-cli/cmd/config"
	"github.com/isai-arellano/kolyn-cli/cmd/ui"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Gestiona la configuración global de Kolyn",
	Long:  `Permite ver y modificar la configuración global almacenada en ~/.kolyn/config.json`,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Inicia el asistente de configuración global (Zero Config)",
	Long:  `Te guía paso a paso para configurar idioma, repositorio de skills y preferencias globales.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runConfigInit(cmd.Context())
	},
}

func init() {
	configCmd.AddCommand(configInitCmd)
}

func runConfigInit(ctx context.Context) error {
	ui.ShowSection("⚙️  Kolyn Global Config")

	ui.PrintStep("Bienvenido al asistente de configuración global.")
	ui.Gray.Println("Esta configuración se guardará en ~/.kolyn/config.json y se usará por defecto en todos tus proyectos.")
	fmt.Println()

	// 1. Idioma
	lang := ui.SelectLanguage()
	ui.CurrentLanguage = lang

	// 2. Repo de Skills
	ui.PrintQuestion(ui.GetText("skills_repo_prompt", "Ingresa la URL del repositorio de skills de tu equipo (ej. git@github.com:org/skills.git):"))
	repoURL := ui.ReadInput("> ")

	var sources []string
	if repoURL != "" {
		sources = []string{repoURL}
	} else {
		// No default repo provided
		sources = []string{}
		ui.PrintInfo("No se configuró repositorio de skills. Usa 'kolyn config' para agregarlo después.")
	}

	// 3. Guardar
	cfg := &config.GlobalConfig{
		Language:      lang,
		SkillsSources: sources,
	}

	if err := config.SaveGlobalConfig(cfg); err != nil {
		return fmt.Errorf("error guardando configuración: %w", err)
	}

	ui.PrintSuccess(ui.GetText("global_created"))
	ui.Gray.Println("Ahora puedes ejecutar 'kolyn sync' en cualquier proyecto para descargar estas skills.")

	return nil
}
