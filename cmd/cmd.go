package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bench-toolset",
	Short: "toolset for benchmark",
}

func Execute() {
	rootCmd.SetOut(os.Stdout)
	if err := rootCmd.Execute(); err != nil {
		rootCmd.Println(err)
		os.Exit(1)
	}
}
