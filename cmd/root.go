package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/danielmesquitta/prisma-to-go/internal/usecase"
	"github.com/spf13/cobra"
)

var schemaFile string = "./schema.prisma"
var outDir string = "./models"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "prisma-to-go",
	Short: "Convert schema.prisma to Go structs",
	Long:  `Convert schema.prisma to Go structs`,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate that required flags are provided
		if schemaFile == "" || outDir == "" {
			fmt.Println("Error: Both --schema and --out are required.")
			_ = cmd.Usage() // Show usage info on error
			os.Exit(1)
		}

		outFile, err := usecase.ParsePrismaSchemaToGoStructs(schemaFile, outDir)
		if err != nil {
			fmt.Println("Error parsing file:", err)
			os.Exit(1)
		}

		command := exec.Command("gofmt", "-w", outFile)
		if err := command.Run(); err != nil {
			fmt.Println("Error running gofmt:", err)
			os.Exit(1)
		}

		fmt.Println("Successfully parsed prisma schema to Golang structs")
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
	var schemaFlag, outFlag string

	rootCmd.Flags().
		StringVarP(&schemaFlag, "schema", "s", "", "Path to the Prisma schema file (default: ./schema.prisma)")
	rootCmd.Flags().
		StringVarP(&outFlag, "out", "o", "", "Output directory for Go structs (default: ./models)")

	if schemaFlag != "" {
		schemaFile = schemaFlag
	}
	if outFlag != "" {
		outDir = outFlag
	}
}
