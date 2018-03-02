package sqlite

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func Connect() (db *sql.DB, err error) {
	db, err = sql.Open("sqlite3", "./log-analyse.sqlite")
	if err != nil {
		return
	}
	err = createTable(db, `CREATE TABLE IF NOT EXISTS log_file (
		uuid TEXT PRIMARY KEY,
		path TEXT NOT NULL,
		no_lines INTEGER NOT NULL,
		start_time TIMESTAMP NOT NULL,
		end_time TIMESTAMP NULL)`)

	err = createTable(db, `CREATE TABLE IF NOT EXISTS log_line (
		uuid TEXT PRIMARY KEY,
		line_no INTEGER NOT NULL,
		value TEXT NOT NULL,
		start_time TIMESTAMP NOT NULL,
		log_file_uuid TEXT NOT NULL,
		FOREIGN KEY (log_file_uuid) REFERENCES log_file(uuid))`)
	err = createTable(db, `CREATE TABLE IF NOT EXISTS log_field (
		uuid TEXT PRIMARY KEY,
		field_type TEXT NOT NULL,
		value_type TEXT NOT NULL,
		value TEXT NOT NULL,
		start_time TIMESTAMP NOT NULL,
		log_line_uuid TEXT NOT NULL,
		FOREIGN KEY (log_line_uuid) REFERENCES log_line(uuid))`)
	return
}

func createTable(db *sql.DB, query string) (err error) {
	stmt, err := db.Prepare(query)
	if err == nil {
		_, err = stmt.Exec()
	}
	return
}
