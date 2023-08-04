package database

import (
	"database/sql"
	"fmt"
	"github.com/rs/zerolog/log"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
)

func InitDatabaseConnection(dbFile string, sqlHistoryDirectory string) (*sql.DB, error) {
	if !fileExists(dbFile) {
		return nil, &DbNotFoundException{message: "Файл БД по пути " + dbFile + " не найден"}
	}
	sqlDB, err := sql.Open("sqlite", dbFile)
	if err == nil {
		err = sqlDB.Ping()
	}

	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(sqlHistoryDirectory)
	if err != nil {
		return nil, err
	}

	log.Info().Msg("Соединение с базой данных успешно установлено")

	// Execute each SQL file
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".sql" {
			content, err := os.ReadFile(filepath.Join(sqlHistoryDirectory, f.Name()))
			if err != nil {
				return nil, err
			}

			_, err = sqlDB.Exec(string(content))
			if err != nil {
				return nil, fmt.Errorf("error executing SQL file %s: %v", f.Name(), err)
			}
		}
	}

	return sqlDB, err
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

type DbNotFoundException struct {
	message string
}

func (e *DbNotFoundException) Error() string {
	return e.message
}
