package cmd

import (
	"bufio"
	"fmt"
	"strings"
)

// readInput lee una línea del reader proporcionado y la limpia.
func readInput(r *bufio.Reader) (string, error) {
	input, err := r.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error leyendo entrada: %w", err)
	}
	return strings.TrimSpace(input), nil
}
