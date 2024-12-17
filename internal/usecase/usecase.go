package usecase

import (
	"fmt"
	"os"
	"os/exec"
)

// Maps Prisma types to Go types
var typeMap = map[string]string{
	"BigInt":   "int64",
	"Boolean":  "bool",
	"Bytes":    "[]byte",
	"DateTime": "time.Time",
	"Decimal":  "float64",
	"Float":    "float64",
	"Int":      "int",
	"String":   "string",
	"Json":     "string",
}

// writeToFile writes the given content to a file
func writeToFile(outDir, filePath, content string) error {
	if err := os.MkdirAll(outDir, os.ModePerm); err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}

	return os.WriteFile(filePath, []byte(content), 0644)
}

// formatGoFile formats the given go file
func formatGoFile(filePath string) error {
	command := exec.Command("gofmt", "-w", filePath)
	return command.Run()
}
