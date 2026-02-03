package ui

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

// SkillOption represents a skill in the selection list
type SkillOption struct {
	Label       string // Display name (e.g. "Web › Next.js")
	Value       string // Unique identifier (e.g. file path)
	Description string // Optional description shown in help text
	Selected    bool   // Initial state
}

// SelectSkills displays an interactive multi-select list using 'huh'.
func SelectSkills(prompt string, options []SkillOption) ([]string, error) {
	if len(options) == 0 {
		return []string{}, nil
	}

	// Build huh options
	var formOptions []huh.Option[string]
	var defaults []string

	for _, opt := range options {
		hOpt := huh.NewOption(opt.Label, opt.Value)
		formOptions = append(formOptions, hOpt)

		if opt.Selected {
			defaults = append(defaults, opt.Value)
		}
	}

	var selectedValues []string

	// Create and run the form
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title(prompt).
				Description("Space to select • Enter to confirm • Arrows to navigate").
				Options(formOptions...).
				Value(&selectedValues).
				Height(15).       // Fixed height for better visibility
				Filterable(true). // Allow searching!
				Limit(10),        // Limit visible items
		),
	).WithTheme(huh.ThemeCatppuccin()) // Modern theme

	// Pre-select defaults if any (Huh handles defaults via Value pointer usually,
	// but for MultiSelect we might need to verify if passing pre-filled slice works.
	// Documentation says 'Value' takes a pointer to the slice where results will be stored.
	// If that slice is pre-populated, those are selected.)
	selectedValues = defaults

	err := form.Run()
	if err != nil {
		return nil, fmt.Errorf("selection cancelled or failed: %w", err)
	}

	return selectedValues, nil
}
