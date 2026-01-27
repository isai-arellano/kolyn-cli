//go:build windows

package cmd

import (
	"os"
	"os/exec"
)

// replaceProcess starts the command detached and exits the current process.
// Windows doesn't support syscall.Exec in the same way.
func replaceProcess(bin string, args []string) error {
	cmd := exec.Command(bin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return err
	}

	os.Exit(0)
	return nil
}
