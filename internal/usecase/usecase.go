package usecase

import (
	"fmt"
	"os"
	"os/exec"
)

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
