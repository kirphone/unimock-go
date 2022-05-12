package database

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func InitDatabaseConnection() (*sql.DB, error) {
	sqlDB, err := sql.Open("sqlite3", "database/unimock.db")
	if err == nil {
		zap.L().Info("Соединение с базой данных успешно установлено")
	} else {
		zap.L().Error(fmt.Sprintf("При соединении с базой данных произошла ошибка: %s", err.Error()))
	}

	return sqlDB, err
}
