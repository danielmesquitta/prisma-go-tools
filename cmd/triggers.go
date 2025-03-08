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

var triggersSchemaFile string

// entitiesCmd represents the triggers command
var triggersCmd = &cobra.Command{
	Use:   "triggers",
	Short: "Create PostgreSQL updated at triggers from schema.prisma files",
	Long:  `Create PostgreSQL updated at triggers from schema.prisma files.`,
	Run: func(cmd *cobra.Command, args []string) {
		outsFiles, err := usecase.CreateUpdatedAtTriggers(triggersSchemaFile)
		if err != nil {
			fmt.Println("prisma-go-tools: ", err)
			os.Exit(1)
		}

		for _, outFile := range outsFiles {
			fmt.Printf("prisma-go-tools triggers: wrote %s\n", outFile)
		}
	},
}

func init() {
	rootCmd.AddCommand(triggersCmd)
	triggersCmd.Flags().
		StringVarP(&triggersSchemaFile, "schema", "s", "./schema.prisma", "Path to the Prisma schema file")
}
