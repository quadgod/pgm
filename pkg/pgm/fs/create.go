package fs

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/quadgod/pgm/pkg/pgm"
)

func touch(fileName string) error {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		file, err := os.Create(fileName)
		if err != nil {
			return err
		}
		defer func(file *os.File) {
			_ = file.Close()
		}(file)
	} else {
		currentTime := time.Now().Local()
		err = os.Chtimes(fileName, currentTime, currentTime)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateMigration создает файлы up & down миграции
func CreateMigration(migDir string, migName string) (*pgm.Migration, error) {
	err := os.MkdirAll(migDir, 0755)
	if err != nil {
		return nil, err
	}

	now := time.Now().UnixMilli()
	upMigrationFile := path.Join(migDir, fmt.Sprintf("%d_%s.up.sql", now, migName))
	downMigrationFile := path.Join(migDir, fmt.Sprintf("%d_%s.down.sql", now, migName))

	migration := new(pgm.Migration)
	migration.Name = fmt.Sprintf("%d_%s", now, migName)

	err = touch(upMigrationFile)
	if err != nil {
		return nil, err
	}

	migration.Up = upMigrationFile

	err = touch(downMigrationFile)
	if err != nil {
		return migration, nil
	}

	migration.Down = downMigrationFile

	return migration, nil
}
