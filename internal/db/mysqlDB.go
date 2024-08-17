package db

import (
	"database/sql"
	"fmt"
	"os"
)

type PersistentFactoryInterface interface {
	CreateMySQLDatabase() (*sql.DB, error)
}

type PersistentFactory struct{}

func (factory *PersistentFactory) CreateMySQLDatabase() (*sql.DB, error) {
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	// Create table and index if not exists
	err = createTableAndIndex(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func createTableAndIndex(db *sql.DB) error {
	// Create users table if not exists
	createTableQuery := `CREATE TABLE IF NOT EXISTS users (
        id INT AUTO_INCREMENT PRIMARY KEY,
        username VARCHAR(255) UNIQUE NOT NULL,
        email VARCHAR(255) UNIQUE NOT NULL,
        first_name VARCHAR(255),
        last_name VARCHAR(255),
        birthday DATE,
        password VARCHAR(255) NOT NULL
    );`
	_, err := db.Exec(createTableQuery)
	if err != nil {
		return fmt.Errorf("error creating table: %v", err)
	}

	// Optionally, create indices or other necessary structures
	// Add full-text index on username if needed
	createIndexQuery := `ALTER TABLE users ADD FULLTEXT(username);`
	_, err = db.Exec(createIndexQuery)
	if err != nil {
		return fmt.Errorf("error creating full-text index: %v", err)
	}

	return nil
}
