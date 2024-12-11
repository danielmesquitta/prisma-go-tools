package cmd

import (
	"fmt"
	"os"

	"github.com/danielmesquitta/prisma-to-go/internal/usecase"
	"github.com/spf13/cobra"
)

var schemaFile, outDir string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "prisma-to-go",
	Short: "Convert schema.prisma to Go structs",
	Long:  `Convert schema.prisma to Go structs`,
	Run: func(cmd *cobra.Command, args []string) {
		outFile, err := usecase.ParsePrismaSchemaToGoStructs(schemaFile, outDir)
		if err != nil {
			fmt.Println("prisma-to-go: ", err)
			os.Exit(1)
		}

		fmt.Printf("prisma-to-go: wrote %s\n", outFile)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().
		StringVarP(&schemaFile, "schema", "s", "./schema.prisma", "Path to the Prisma schema file (default: ./schema.prisma)")
	rootCmd.Flags().
		StringVarP(&outDir, "output", "o", "./models", "Output directory for Go structs (default: ./models)")
}
