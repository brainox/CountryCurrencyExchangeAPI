package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "api.db")

	if err != nil {
		panic("Could not connect to database.")
	}

	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(5)

	createTables()
}

func createTables() {
	createCountryTable := `CREATE TABLE IF NOT EXISTS countries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		capital TEXT,
		region TEXT,
		population INTEGER NOT NULL,
		currency_code TEXT,
		exchange_rate REAL NOT NULL,
		estimated_gdp REAL,
		flag_url TEXT,
		last_refreshed_at TEXT
	);`
	_, err := DB.Exec(createCountryTable)
	if err != nil {
		panic("Could not create countries table.")
	}
}
