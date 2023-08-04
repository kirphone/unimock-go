package database

import (
	"database/sql"
	"fmt"
	"github.com/rs/zerolog/log"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
)

const UpdateVersionQuery = "UPDATE version SET version = ?"
const SelectVersionTableQuery = "SELECT name FROM sqlite_master WHERE type='table' AND name='version'"

const SelectVersionQuery = "SELECT version FROM version LIMIT 1"

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

	dbVersion := -1

	row := sqlDB.QueryRow(SelectVersionTableQuery)
	var name string
	err = row.Scan(&name)

	if err == nil && name == "version" {
		row = sqlDB.QueryRow(SelectVersionQuery)
		err = row.Scan(&dbVersion)
		if err != nil {
			log.Error().Err(err).Msg("Ошибка при получении версии базы данных")
			dbVersion = -1
		}
	}

	err = nil

	// Execute each SQL file
	for i := dbVersion + 1; i < len(files); i++ {
		if filepath.Ext(files[i].Name()) == ".sql" {
			content, err := os.ReadFile(filepath.Join(sqlHistoryDirectory, files[i].Name()))
			if err != nil {
				return nil, err
			}

			_, err = sqlDB.Exec(string(content))
			if err != nil {
				return nil, fmt.Errorf("error executing SQL file %s: %v", files[i].Name(), err)
			}

			// Update version in database
			_, err = sqlDB.Exec(UpdateVersionQuery, i)
			if err != nil {
				return nil, fmt.Errorf("error updating database version to %d: %v", i, err)
			}

			dbVersion = i
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
