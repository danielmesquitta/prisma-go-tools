package usecase

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Model represents a Prisma model and its fields.
type Model struct {
	ModelName string
	TableName string
	Fields    []Field
}

// Field represents a single field in a Prisma model.
type Field struct {
	FieldName    string
	ColumnName   string
	HasUpdatedAt bool
}

func CreateUpdatedAtTriggers(
	schemaPath string,
) ([]string, error) {
	migrationsDir := filepath.Join(filepath.Dir(schemaPath), "migrations")

	models, err := parseSchemaPrisma(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("error parsing schema.prisma: %w", err)
	}

	existsUpdatedAtTriggersByTableName, err := findExistingUpdatedAtTriggersByTableName(
		migrationsDir,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"error finding existing updated at triggers: %w",
			err,
		)
	}

	migrationFiles := []string{}
	for _, model := range models {
		var updatedAtCols []string
		for _, field := range model.Fields {
			if field.HasUpdatedAt {
				updatedAtCols = append(updatedAtCols, field.ColumnName)
			}
		}

		if len(updatedAtCols) == 0 {
			continue
		}

		if _, ok := existsUpdatedAtTriggersByTableName[model.TableName]; ok {
			continue
		}

		triggerSQL := generateTriggerSQL(model.TableName, updatedAtCols)

		migrationFile, err := createNewMigrationFile(
			migrationsDir,
			model.TableName,
			triggerSQL,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"error creating new migration for table %s: %w",
				model.TableName,
				err,
			)
		}

		migrationFiles = append(migrationFiles, migrationFile)
	}

	return migrationFiles, nil
}

// parseSchemaPrisma reads the `schema.prisma` file and extracts:
//   - model name
//   - table name (@@map if present, otherwise model name)
//   - fields that have @updatedAt
//   - column name (@map if present, otherwise field name)
func parseSchemaPrisma(schemaPath string) ([]Model, error) {
	file, err := os.Open(schemaPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var models []Model

	scanner := bufio.NewScanner(file)

	var currentModel *Model
	var inModelBlock bool

	// Regex to match lines like:
	// model User {
	// model SomethingElse {
	modelStartRegex := regexp.MustCompile(`^model\s+(\w+)\s+\{`)

	// Regex to match lines like:
	// @@map("users")
	mapTableRegex := regexp.MustCompile(`@@map\("([^"]+)"\)`)

	// Regex to match fields, e.g.:
	// updatedAt     DateTime   @updatedAt @map("updated_at")
	// lastModified DateTime   @updatedAt
	fieldRegex := regexp.MustCompile(`^(\w+)\s+(\w+)\s+(.+)$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if matches := modelStartRegex.FindStringSubmatch(line); len(
			matches,
		) == 2 {
			inModelBlock = true
			currentModel = &Model{
				ModelName: matches[1],
				TableName: matches[1],
			}
			continue
		}

		if inModelBlock && strings.HasPrefix(line, "}") {
			if currentModel != nil {
				models = append(models, *currentModel)
				currentModel = nil
			}
			inModelBlock = false
			continue
		}

		if inModelBlock && currentModel != nil {
			if mapMatches := mapTableRegex.FindStringSubmatch(line); len(
				mapMatches,
			) == 2 {
				currentModel.TableName = mapMatches[1]
				continue
			}

			if fieldMatches := fieldRegex.FindStringSubmatch(line); len(
				fieldMatches,
			) == 4 {
				fieldName := fieldMatches[1]
				// fieldType := fieldMatches[2]
				fieldAttrs := fieldMatches[3]

				hasUpdatedAt := strings.Contains(fieldAttrs, "@updatedAt")

				var columnName string
				columnMapRegex := regexp.MustCompile(`@map\("([^"]+)"\)`)
				if colMapMatch := columnMapRegex.FindStringSubmatch(fieldAttrs); len(
					colMapMatch,
				) == 2 {
					columnName = colMapMatch[1]
				} else {
					columnName = fieldName
				}

				if hasUpdatedAt {
					currentModel.Fields = append(currentModel.Fields, Field{
						FieldName:    fieldName,
						ColumnName:   columnName,
						HasUpdatedAt: true,
					})
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return models, nil
}

// findExistingUpdatedAtTriggersByTableName scans the prisma/migrations folder
// to find all existing updated_at triggers.
func findExistingUpdatedAtTriggersByTableName(
	migrationsDir string,
) (map[string]struct{}, error) {
	existingTriggers := map[string]struct{}{}

	err := filepath.Walk(
		migrationsDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				return filepath.SkipDir
			}
			split := strings.Split(info.Name(), "_updated_at_")
			if len(split) != 2 {
				return nil
			}
			tableName := split[1]
			existingTriggers[tableName] = struct{}{}
			return nil
		},
	)
	if !(err == filepath.SkipDir || err == nil) {
		return nil, err
	}

	return existingTriggers, nil
}

// generateTriggerSQL generates a block of SQL that creates a trigger to auto-update
// any @updatedAt column in the specified table. Typically, you only need one
// trigger function per table that sets all the @updatedAt columns to NOW().
func generateTriggerSQL(tableName string, columns []string) string {
	setClauses := make([]string, 0, len(columns))
	for _, col := range columns {
		setClauses = append(setClauses, fmt.Sprintf("NEW.\"%s\" = now();", col))
	}

	return fmt.Sprintf(`
-- Auto-generated trigger for table "%[1]s" to update @updatedAt columns
CREATE OR REPLACE FUNCTION "%[1]s_updated_at_trigger"()
RETURNS TRIGGER AS $$
BEGIN
    %[2]s
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER "%[1]s_updated_at_trigger"
BEFORE UPDATE ON "%[1]s"
FOR EACH ROW
EXECUTE PROCEDURE "%[1]s_updated_at_trigger"();
`, tableName, strings.Join(setClauses, "\n    "))
}

func createNewMigrationFile(
	migrationsDir, tableName, sqlStmt string,
) (string, error) {
	timestamp := time.Now().Format("20060102150405")
	migrationName := fmt.Sprintf("%s_updated_at_%s", timestamp, tableName)
	migrationFolder := filepath.Join(migrationsDir, migrationName)

	if err := os.MkdirAll(migrationFolder, 0o755); err != nil {
		return "", err
	}

	migrationFilePath := filepath.Join(migrationFolder, "migration.sql")

	err := os.WriteFile(migrationFilePath, []byte(sqlStmt), 0o644)
	if err != nil {
		return "", err
	}

	return migrationFilePath, nil
}
