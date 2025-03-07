package usecase

import (
	"os"
	"path/filepath"
	"strings"
)

func UnZipMigrations(
	schemaPath string,
) error {
	migrationsDir := filepath.Join(filepath.Dir(schemaPath), "migrations")

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dirName := entry.Name()
		if !strings.Contains(dirName, "_") {
			continue
		}

		srcPath := filepath.Join(migrationsDir, dirName, "migration.sql")

		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			continue
		}

		destPath := filepath.Join(migrationsDir, dirName+".sql")

		if err := copyFile(srcPath, destPath); err != nil {
			continue
		}

		if err := os.RemoveAll(filepath.Join(migrationsDir, dirName)); err != nil {
			continue
		}
	}

	return nil
}
