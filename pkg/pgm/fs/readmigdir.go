package fs

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/quadgod/pgm/pkg/pgm"
)

// group группирует отсортированные sql файлы в up & down пары
func group(sortedFiles []string, migDir string) ([]pgm.Migration, error) {
	migrations := make([]pgm.Migration, 0)

	if len(sortedFiles) == 0 {
		return migrations, nil
	}

	for i := 0; i < len(sortedFiles); i += 2 {
		if strings.HasSuffix(sortedFiles[i], ".up.sql") {
			return nil, fmt.Errorf("not found down migration for \"%s\" migration", sortedFiles[i])
		}

		downName, found := strings.CutSuffix(sortedFiles[i], ".down.sql")
		if !found {
			return nil, fmt.Errorf("invalid migration filename \"%s\"", sortedFiles[i])
		}

		if i+1 >= len(sortedFiles) {
			return nil, fmt.Errorf("not found up migration for \"%s\" migration", sortedFiles[i])
		}

		upName, found := strings.CutSuffix(sortedFiles[i+1], ".up.sql")
		if !found {
			return nil, fmt.Errorf("not found up migration for \"%s\" migration", sortedFiles[i])
		}

		if downName != upName {
			return nil, fmt.Errorf("not found up migration for \"%s\" migration", sortedFiles[i])
		}

		migration := new(pgm.Migration)
		migration.Name = downName
		migration.Down = path.Join(migDir, sortedFiles[i])
		migration.Up = path.Join(migDir, sortedFiles[i+1])
		migrations = append(migrations, *migration)
	}

	return migrations, nil
}

// ReadMigrationsDir читает список миграций из директории
func ReadMigrationsDir(migDir string) ([]pgm.Migration, error) {
	if _, err := os.Stat(migDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("migrations directory does not exist: %w", err)
	}

	dirEntries, err := os.ReadDir(migDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory %q: %w", migDir, err)
	}

	fNames := make([]string, 0)
	for _, f := range dirEntries {
		if f.IsDir() {
			continue
		}

		if !strings.HasSuffix(f.Name(), ".up.sql") && !strings.HasSuffix(f.Name(), ".down.sql") {
			continue
		}

		filenameRegexp, err := regexp.Compile(`^[a-zA-Z0-9_.]+$`)
		if err != nil {
			return nil, err
		}

		match := filenameRegexp.MatchString(f.Name())
		if !match {
			return nil, fmt.Errorf("invalid migration file name found \"%s\"", f.Name())
		}

		fNames = append(fNames, f.Name())
	}

	if len(fNames) != 0 {
		sort.Strings(fNames)
	}

	migrations, err := group(fNames, migDir)
	if err != nil {
		return nil, err
	}

	return migrations, nil
}
