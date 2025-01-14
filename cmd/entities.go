/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/danielmesquitta/prisma-to-go/internal/usecase"
	"github.com/spf13/cobra"
)

var entitiesSchemaFile, entitiesOutDir string

// entitiesCmd represents the entities command
var entitiesCmd = &cobra.Command{
	Use:   "entities",
	Short: "Convert schema.prisma models to Go structs",
	Long:  `Convert schema.prisma models to Go structs.`,
	Run: func(cmd *cobra.Command, args []string) {
		outFile, err := usecase.ParsePrismaSchemaToGoStructs(
			entitiesSchemaFile,
			entitiesOutDir,
		)
		if err != nil {
			fmt.Println("prisma-to-go: ", err)
			os.Exit(1)
		}

		fmt.Printf("prisma-to-go entities: wrote %s\n", outFile)
	},
}

func init() {
	rootCmd.AddCommand(entitiesCmd)
	entitiesCmd.Flags().
		StringVarP(&entitiesSchemaFile, "schema", "s", "./schema.prisma", "Path to the Prisma schema file")
	entitiesCmd.Flags().
		StringVarP(&entitiesOutDir, "output", "o", "./models", "Output directory for Go entities structs")
}
