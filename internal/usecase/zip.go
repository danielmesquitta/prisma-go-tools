package usecase

import (
	"os"
	"path/filepath"
	"strings"
)

func ZipMigrations(
	schemaPath string,
) error {
	migrationsDir := filepath.Join(filepath.Dir(schemaPath), "migrations")

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		if !strings.HasSuffix(fileName, ".sql") {
			continue
		}

		dirName := strings.TrimSuffix(fileName, ".sql")

		newDir := filepath.Join(migrationsDir, dirName)
		if err := os.MkdirAll(newDir, 0755); err != nil {
			continue
		}

		srcPath := filepath.Join(migrationsDir, fileName)
		destPath := filepath.Join(newDir, "migration.sql")

		if err := copyFile(srcPath, destPath); err != nil {
			// Clean up the directory if copy fails
			os.RemoveAll(newDir)
			continue
		}

		if err := os.Remove(srcPath); err != nil {
			continue
		}

	}

	return nil
}
