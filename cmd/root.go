package cmd

import (
	"os"
	"time"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "prisma-to-go",
	Short: "Convert schema.prisma to Go structs",
	Long:  `Convert schema.prisma to Go structs`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	time.Local = time.UTC

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
