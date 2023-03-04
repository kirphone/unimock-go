package database

import (
	"database/sql"
	"fmt"
	"go.uber.org/zap"
	_ "modernc.org/sqlite"
)

func InitDatabaseConnection() (*sql.DB, error) {
	sqlDB, err := sql.Open("sqlite", "database/unimock.db")
	if err == nil {
		zap.L().Info("Соединение с базой данных успешно установлено")
	} else {
		zap.L().Error(fmt.Sprintf("При соединении с базой данных произошла ошибка: %s", err.Error()))
	}

	return sqlDB, err
}
