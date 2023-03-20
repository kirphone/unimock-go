package database

import (
	"database/sql"
	_ "modernc.org/sqlite"
	"os"
)

func InitDatabaseConnection(dbFile string) (*sql.DB, error) {
	if !fileExists(dbFile) {
		return nil, &DbNotFoundException{message: "Файл БД по пути " + dbFile + " не найден"}
	}
	sqlDB, err := sql.Open("sqlite", dbFile)
	if err == nil {
		err = sqlDB.Ping()
	}

	return sqlDB, err
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

type DbNotFoundException struct {
	message string
}

func (e *DbNotFoundException) Error() string {
	return e.message
}
