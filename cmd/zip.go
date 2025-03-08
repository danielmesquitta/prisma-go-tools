package cmd

import (
	"fmt"
	"os"

	"github.com/danielmesquitta/prisma-go-tools/internal/usecase"
	"github.com/spf13/cobra"
)

var zipSchemaFile string

// zipCmd represents the zip command
var zipCmd = &cobra.Command{
	Use:   "zip",
	Short: "Zip migrations dir",
	Long:  `Zip migrations dir`,
	Run: func(cmd *cobra.Command, args []string) {
		err := usecase.UnZipMigrations(zipSchemaFile)
		if err != nil {
			fmt.Println("prisma-go-tools: ", err)
			os.Exit(1)
		}

		fmt.Printf("prisma-go-tools zip: done\n")
	},
}

func init() {
	rootCmd.AddCommand(zipCmd)
	zipCmd.Flags().
		StringVarP(&zipSchemaFile, "schema", "s", "./schema.prisma", "Path to the Prisma schema file")
}
