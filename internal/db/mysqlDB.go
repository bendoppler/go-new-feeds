package db

import (
	"database/sql"
	"fmt"
	"news-feed/pkg/config/webApp"
	"time"
)

type PersistentFactoryInterface interface {
	CreateMySQLDatabase() (*sql.DB, error)
}

type PersistentFactory struct{}

func (factory *PersistentFactory) CreateMySQLDatabase() (*sql.DB, error) {
	cfg := webApp.LoadConfig()
	user := cfg.DBUser
	password := cfg.DBPassword
	host := cfg.DBHost
	port := cfg.DBPort
	dbName := cfg.DBName

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, password, host, port, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	// Create table and index if not exists
	err = Migrate(db)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(100)                // Max number of open connections
	db.SetMaxIdleConns(5)                  // Max number of idle connections
	db.SetConnMaxLifetime(time.Second * 5) // Recycle connections periodically

	return db, nil
}
