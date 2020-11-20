package cmd

import (
	"os"
	"os/exec"
)

func runBrRestore(args []string) error {
	args = append([]string{"restore"}, args...)
	cmd := exec.Command("br", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
