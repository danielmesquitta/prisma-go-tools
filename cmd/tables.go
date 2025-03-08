/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/danielmesquitta/prisma-go-tools/internal/usecase"
	"github.com/spf13/cobra"
)

var tablesSchemaFile, tablesOutDir string

// tablesCmd represents the tables command
var tablesCmd = &cobra.Command{
	Use:   "tables",
	Short: "Convert schema.prisma tables to a Go custom type",
	Long:  `Convert schema.prisma tables to a Go custom type`,
	Run: func(cmd *cobra.Command, args []string) {
		outFile, err := usecase.ParsePrismaTables(
			tablesSchemaFile,
			tablesOutDir,
		)
		if err != nil {
			fmt.Println("prisma-go-tools: ", err)
			os.Exit(1)
		}

		fmt.Printf("prisma-go-tools tables: wrote %s\n", outFile)
	},
}

func init() {
	rootCmd.AddCommand(tablesCmd)
	tablesCmd.Flags().
		StringVarP(&tablesSchemaFile, "schema", "s", "./schema.prisma", "Path to the Prisma schema file")
	tablesCmd.Flags().
		StringVarP(&tablesOutDir, "output", "o", "./tables", "Output directory for Go Table custom type")
}
