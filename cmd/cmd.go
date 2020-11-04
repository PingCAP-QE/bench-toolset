package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "stability_bench",
	Short: "toolset for stability benchmark",
}

func Execute() {
	rootCmd.SetOut(os.Stdout)
	if err := rootCmd.Execute(); err != nil {
		rootCmd.Println(err)
		os.Exit(1)
	}
}
