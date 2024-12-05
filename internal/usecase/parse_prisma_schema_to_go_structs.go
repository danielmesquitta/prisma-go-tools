package usecase

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/iancoleman/strcase"
)

// Maps Prisma types to Go types
var typeMap = map[string]string{
	"String":   "string",
	"DateTime": "time.Time",
	"Int":      "int",
	"Float":    "float64",
	"BigInt":   "int64",
	"Decimal":  "float64",
	"Json":     "string",
	"Boolean":  "bool",
	"Bytes":    "[]byte",
}

func ParsePrismaSchemaToGoStructs(schemaPath, outDir string) error {
	return processSchema(schemaPath, outDir)
}

// Parse a Prisma model into a Go struct
func parseModel(lines []string) (string, bool, bool) {
	structName := ""
	fields := []string{}
	modelRegex := regexp.MustCompile(`model\s+(\w+)`)
	fieldRegex := regexp.MustCompile(`\s*(\w+)\s+(\w+)(\[\])?\s*(\?)?.*`)
	uuidRegex := regexp.MustCompile(`@db\.Uuid`)

	usesTime := false
	usesUUID := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if modelMatch := modelRegex.FindStringSubmatch(line); modelMatch != nil {
			structName = modelMatch[1]
		} else if fieldMatch := fieldRegex.FindStringSubmatch(line); fieldMatch != nil {
			fieldName := strcase.ToCamel(fieldMatch[1])

			if strings.HasSuffix(fieldName, "Id") {
				fieldName = strings.TrimSuffix(fieldName, "Id")
				fieldName += "ID"
			}

			fieldType := typeMap[fieldMatch[2]]
			if uuidRegex.MatchString(line) {
				fieldType = "uuid.UUID"
				usesUUID = true
			}

			// Handle enums or custom types
			if _, ok := typeMap[fieldMatch[2]]; !ok {
				fieldType = fieldMatch[2] // Keep custom type as-is
			}

			// Add list or pointer handling
			if fieldMatch[3] == "[]" {
				fieldType = "[]" + fieldType // Only apply once for list types
			} else if fieldMatch[4] == "?" {
				fieldType = "*" + fieldType
			}

			if fieldType == "time.Time" {
				usesTime = true
			}

			fields = append(fields, fmt.Sprintf("\t%s %s `json:\"%s\"`", fieldName, fieldType, fieldMatch[1]))
		}
	}

	structDefinition := fmt.Sprintf(
		"type %s struct {\n%s\n}",
		structName,
		strings.Join(fields, "\n"),
	)
	return structDefinition, usesTime, usesUUID
}

// Parse a Prisma enum into a Go type
func parseEnum(lines []string) string {
	enumName := ""
	values := []string{}
	enumRegex := regexp.MustCompile(`enum\s+(\w+)`)
	valueRegex := regexp.MustCompile(`^\s*(\w+)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if enumMatch := enumRegex.FindStringSubmatch(line); enumMatch != nil {
			enumName = enumMatch[1]
		} else if valueMatch := valueRegex.FindStringSubmatch(line); valueMatch != nil {
			value := valueMatch[1]
			values = append(values, value)
		}
	}

	// Generate Go enum type and constants
	var enumDef strings.Builder
	enumDef.WriteString(fmt.Sprintf("type %s string\n\nconst (\n", enumName))
	for _, value := range values {
		enumDef.WriteString(
			fmt.Sprintf(
				"\t%s%s %s = \"%s\"\n",
				enumName,
				strcase.ToCamel(value),
				enumName,
				value,
			),
		)
	}
	enumDef.WriteString(")\n")

	return enumDef.String()
}

// Reads and processes the Prisma schema file
func processSchema(filePath, outDir string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	var lines []string
	var blocks [][]string
	var result strings.Builder
	scanner := bufio.NewScanner(file)
	inBlock := false

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "model ") ||
			strings.HasPrefix(strings.TrimSpace(line), "enum ") {
			if inBlock {
				blocks = append(blocks, lines)
			}
			inBlock = true
			lines = []string{line}
		} else if inBlock {
			lines = append(lines, line)
		}
	}
	if inBlock {
		blocks = append(blocks, lines)
	}

	usesTime := false
	usesUUID := false

	for _, block := range blocks {
		if strings.HasPrefix(strings.TrimSpace(block[0]), "model ") {
			structDef, timeUsed, uuidUsed := parseModel(block)
			result.WriteString(structDef)
			result.WriteString("\n\n")

			if timeUsed {
				usesTime = true
			}
			if uuidUsed {
				usesUUID = true
			}
		} else if strings.HasPrefix(strings.TrimSpace(block[0]), "enum ") {
			enumDef := parseEnum(block)
			result.WriteString(enumDef)
			result.WriteString("\n\n")
		}
	}

	// Determine package name and output file name
	outDirBase := filepath.Base(outDir)
	outputFileName := filepath.Join(
		outDir,
		fmt.Sprintf("%s_gen.go", outDirBase),
	)

	// Create import block
	imports := ""
	if usesTime || usesUUID {
		imports = "import (\n"
		if usesTime {
			imports += "\t\"time\"\n"
		}
		if usesUUID {
			imports += "\t\"github.com/google/uuid\"\n"
		}
		imports += ")\n\n"
	}

	// Create the full output content
	finalOutput := fmt.Sprintf(
		"//nolint\n//go:build !codeanalysis\n// +build !codeanalysis\n\npackage %s\n\n%s%s",
		outDirBase,
		imports,
		result.String(),
	)

	// Write to the output file
	err = os.MkdirAll(outDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}

	err = os.WriteFile(outputFileName, []byte(finalOutput), 0644)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}
