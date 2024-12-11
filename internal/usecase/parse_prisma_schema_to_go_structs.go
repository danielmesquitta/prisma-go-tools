package usecase

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/danielmesquitta/prisma-to-go/internal/pkg/strcase"
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

func ParsePrismaSchemaToGoStructs(
	schemaPath, outDir string,
) (outFile string, err error) {
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

			fieldType := typeMap[fieldMatch[2]]
			if uuidRegex.MatchString(line) {
				fieldType = "uuid.UUID"
				usesUUID = true
			}

			// Do not include relationships
			if _, ok := typeMap[fieldMatch[2]]; !ok {
				continue
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

			fields = append(fields, fmt.Sprintf("\t%s %s `json:\"%s,omitempty\"`", fieldName, fieldType, fieldMatch[1]))
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

	// Parse enum name and values and add to type map
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if enumMatch := enumRegex.FindStringSubmatch(line); enumMatch != nil {
			enumName = enumMatch[1]
			typeMap[enumName] = enumName
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
func processSchema(filePath, outDir string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening file: %w", err)
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

	// First, parse enums
	for _, block := range blocks {
		if strings.HasPrefix(strings.TrimSpace(block[0]), "enum ") {
			enumDef := parseEnum(block)
			result.WriteString(enumDef)
			result.WriteString("\n\n")
		}
	}

	// Next, parse models
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
		}
	}

	// Determine package name and output file name
	outDirBase := filepath.Base(outDir)
	outputFilePath := filepath.Join(
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
		return "", fmt.Errorf("error creating output directory: %w", err)
	}

	err = os.WriteFile(outputFilePath, []byte(finalOutput), 0644)
	if err != nil {
		return "", fmt.Errorf("error writing to file: %w", err)
	}

	if err := formatGoFile(outputFilePath); err != nil {
		return "", fmt.Errorf("error formatting file: %w", err)
	}

	return outputFilePath, nil
}

func formatGoFile(filePath string) error {
	command := exec.Command("gofmt", "-w", filePath)
	return command.Run()
}
