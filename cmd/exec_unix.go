//go:build !windows

package cmd

import (
	"os"
	"syscall"
)

// replaceProcess replaces the current process with the new command.
// This preserves TTY for sudo prompts and interactive scripts on Unix.
func replaceProcess(bin string, args []string) error {
	// syscall.Exec requires the binary name as the first argument in argv
	argv := append([]string{bin}, args...)
	return syscall.Exec(bin, argv, os.Environ())
}
