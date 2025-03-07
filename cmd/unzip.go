package cmd

import (
	"fmt"
	"os"

	"github.com/danielmesquitta/prisma-to-go/internal/usecase"
	"github.com/spf13/cobra"
)

var unzipSchemaFile string

// unzipCmd represents the unzip command
var unzipCmd = &cobra.Command{
	Use:   "unzip",
	Short: "Unzip migrations dir",
	Long:  `Unzip migrations dir`,
	Run: func(cmd *cobra.Command, args []string) {
		err := usecase.UnZipMigrations(unzipSchemaFile)
		if err != nil {
			fmt.Println("prisma-to-go: ", err)
			os.Exit(1)
		}

		fmt.Printf("prisma-to-go unzip: done!\n")
	},
}

func init() {
	rootCmd.AddCommand(unzipCmd)
	unzipCmd.Flags().
		StringVarP(&unzipSchemaFile, "schema", "s", "./schema.prisma", "Path to the Prisma schema file")
}
