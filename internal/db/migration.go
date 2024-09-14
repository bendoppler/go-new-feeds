package db

import (
	"database/sql"
	"fmt"
)

func Migrate(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS user (
			id INT AUTO_INCREMENT PRIMARY KEY,
			hashed_password VARCHAR(255) NOT NULL,
			salt VARCHAR(255) NOT NULL,
			first_name VARCHAR(255) NOT NULL,
			last_name VARCHAR(255) NOT NULL,
			dob DATE NOT NULL,
			email VARCHAR(255) NOT NULL,
			user_name VARCHAR(255) UNIQUE NOT NULL
		);`,

		`CREATE TABLE IF NOT EXISTS post (
			id INT AUTO_INCREMENT PRIMARY KEY,
			fk_user_id INT NOT NULL,
			content_text TEXT,
			content_image_path VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			visible BOOLEAN DEFAULT TRUE,
			FOREIGN KEY (fk_user_id) REFERENCES user(id)
		);`,

		`CREATE TABLE IF NOT EXISTS comment (
			id INT AUTO_INCREMENT PRIMARY KEY,
			fk_post_id INT NOT NULL,
			fk_user_id INT NOT NULL,
			content TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (fk_post_id) REFERENCES post(id),
			FOREIGN KEY (fk_user_id) REFERENCES user(id)
		);`,

		`CREATE TABLE IF NOT EXISTS like (
			fk_post_id INT NOT NULL,
			fk_user_id INT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (fk_post_id, fk_user_id),
			FOREIGN KEY (fk_post_id) REFERENCES post(id),
			FOREIGN KEY (fk_user_id) REFERENCES user(id)
		);`,

		`CREATE TABLE IF NOT EXISTS user_user (
			fk_user_id INT NOT NULL,
			fk_follower_id INT NOT NULL,
			PRIMARY KEY (fk_user_id, fk_follower_id),
			FOREIGN KEY (fk_user_id) REFERENCES user(id),
			FOREIGN KEY (fk_follower_id) REFERENCES user(id)
		);`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			return fmt.Errorf("error running migration query: %v", err)
		}
	}

	return nil
}
