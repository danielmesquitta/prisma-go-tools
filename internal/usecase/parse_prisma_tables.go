package usecase

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func ParsePrismaTables(
	schemaPath, outDir string,
) (string, error) {
	outputFilePath := filepath.Join(outDir, "table_gen.go")

	// Read the schema.prisma file
	tableNames, err := extractTableNames(schemaPath)
	if err != nil {
		return "", err
	}

	packageName := filepath.Base(outDir)

	// Generate the Go file content
	goFileContent := generateGoFileContent(packageName, tableNames)

	// Write the content to the output Go file
	if err := writeToFile(outDir, outputFilePath, goFileContent); err != nil {
		return "", err
	}

	if err := formatGoFile(outputFilePath); err != nil {
		return "", err
	}

	return outputFilePath, nil
}

// extractTableNames reads the schema.prisma file and extracts table names
func extractTableNames(filePath string) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	tableNames := make(map[string]string) // modelName -> tableName
	scanner := bufio.NewScanner(file)

	modelRegex := regexp.MustCompile(
		`^model\s+(\w+)`,
	) // Matches "model <Name>"
	mapRegex := regexp.MustCompile(
		`@@map\("([^"]+)"\)`,
	) // Matches @@map("table_name")

	var currentModel string
	inModelBlock := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Check if entering a new model
		if matches := modelRegex.FindStringSubmatch(line); len(matches) > 1 {
			currentModel = matches[1]
			inModelBlock = true
			tableNames[currentModel] = strings.ToLower(
				currentModel,
			) // Default to model name
			continue
		}

		// If inside a model block, look for @@map
		if inModelBlock {
			if matches := mapRegex.FindStringSubmatch(line); len(matches) > 1 {
				tableNames[currentModel] = matches[1] // Use mapped table name
			}

			// Check for end of model block
			if strings.HasPrefix(line, "}") {
				inModelBlock = false
				currentModel = ""
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return tableNames, nil
}

// generateGoFileContent generates the content of the Go file
func generateGoFileContent(
	packageName string,
	tableNames map[string]string,
) string {
	var builder strings.Builder

	// Package declaration
	builder.WriteString(fmt.Sprintf("package %s\n\n", packageName))
	builder.WriteString("type Table string\n\n")
	builder.WriteString("const (\n")

	// Write each table name as a constant
	for modelName, tableName := range tableNames {
		constName := "Table" + modelName // e.g., TableCategories
		builder.WriteString(
			fmt.Sprintf("\t%s Table = \"%s\"\n", constName, tableName),
		)
	}

	builder.WriteString(")\n")
	return builder.String()
}
