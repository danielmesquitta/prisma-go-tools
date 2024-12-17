package usecase

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/danielmesquitta/prisma-to-go/internal/pkg/strcase"
)

func ParsePrismaTablesAndColumns(
	schemaPath, outDir string,
) (string, error) {
	outputFilePath := filepath.Join(outDir, "table_gen.go")

	// Extract table names and columns
	tableNames, columns, err := extractTableAndColumnNames(schemaPath)
	if err != nil {
		return "", err
	}

	packageName := filepath.Base(outDir)

	// Generate the Go file content
	goFileContent := generateGoFileContent(packageName, tableNames, columns)

	// Write the content to the output Go file
	if err := writeToFile(outDir, outputFilePath, goFileContent); err != nil {
		return "", err
	}

	if err := formatGoFile(outputFilePath); err != nil {
		return "", err
	}

	return outputFilePath, nil
}

// extractTableAndColumnNames reads schema.prisma and extracts table names and column details
func extractTableAndColumnNames(
	filePath string,
) (map[string]string, map[string]map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	tableNames := make(map[string]string) // modelName -> tableName
	columns := make(
		map[string]map[string]string,
	) // modelName -> {columnName: columnType}
	scanner := bufio.NewScanner(file)

	modelRegex := regexp.MustCompile(`^model\s+(\w+)`)
	mapRegex := regexp.MustCompile(`@@map\("([^"]+)"\)`)
	fieldRegex := regexp.MustCompile(
		`^(\w+)\s+(\w+)`,
	) // Matches columnName columnType

	var currentModel string
	inModelBlock := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Detect new model
		if matches := modelRegex.FindStringSubmatch(line); len(matches) > 1 {
			currentModel = matches[1]
			inModelBlock = true
			tableNames[currentModel] = strings.ToLower(currentModel)
			columns[currentModel] = make(
				map[string]string,
			) // Initialize column map
			continue
		}

		// Parse model fields for column names and types
		if inModelBlock {
			if matches := fieldRegex.FindStringSubmatch(line); len(
				matches,
			) > 2 {
				columnName := matches[1]
				columnType := matches[2]

				// Only add to columns if the column type is in typeMap
				if _, exists := typeMap[columnType]; exists {
					columns[currentModel][columnName] = columnType
				}
			}

			// Parse @@map for table name
			if matches := mapRegex.FindStringSubmatch(line); len(matches) > 1 {
				tableNames[currentModel] = matches[1]
			}

			// Detect end of model block
			if strings.HasPrefix(line, "}") {
				inModelBlock = false
				currentModel = ""
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	return tableNames, columns, nil
}

// generateGoFileContent generates the content of the Go file
func generateGoFileContent(
	packageName string,
	tableNames map[string]string,
	columns map[string]map[string]string,
) string {
	var builder strings.Builder

	// Package declaration
	builder.WriteString(fmt.Sprintf("package %s\n\n", packageName))
	builder.WriteString("type Table string\n")
	builder.WriteString("type Column string\n\n")

	// Table constants
	builder.WriteString("const (\n")
	for modelName, tableName := range tableNames {
		constName := "Table" + modelName
		builder.WriteString(
			fmt.Sprintf("\t%s Table = \"%s\"\n", constName, tableName),
		)
	}
	builder.WriteString(")\n\n")

	// Column constants
	for modelName, cols := range columns {
		builder.WriteString(fmt.Sprintf("// Columns for table %s\n", modelName))
		builder.WriteString("const (\n")
		for colName := range cols {
			constName := fmt.Sprintf(
				"Column%s%s",
				modelName,
				strcase.ToCamel(colName),
			)
			builder.WriteString(
				fmt.Sprintf(
					"\t%s Column = \"%s\"\n",
					constName,
					colName,
				),
			)
		}
		builder.WriteString(")\n\n")
	}

	return builder.String()
}
