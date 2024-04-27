package api

import (
	"database/sql"
	"fmt"
	"log"
	_ "modernc.org/sqlite"
	"os"
)

// createTable - создает таблицу scheduler
func createTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS scheduler (
		    id INTEGER PRIMARY KEY AUTOINCREMENT,
		    date VARCHAR(8) NOT NULL DEFAULT '',
		    title TEXT NOT NULL DEFAULT '',
		    comment TEXT NOT NULL DEFAULT '',
		    repeat VARCHAR(128) 
		    );
		CREATE INDEX scheduler_date ON scheduler(date);
		`)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

// DatabaseCheck DatabaseInit - проверяет наличие базы данных, в случае отсутствия создает ее вместе с таблицей scheduler
func DatabaseCheck() {
	var defaultDbPath = "../go_final_project/scheduler.db"
	var install bool
	// При наличии файла .env c заданным значением переменной TODO_DBFILE база данных будет создана по указанному пути, если переменная не задана, то будет создана в defaultDbPath
	TODO_DBFILE, exists := os.LookupEnv("TODO_DBFILE")
	if exists && TODO_DBFILE != "" {
		_, err := os.Stat(TODO_DBFILE)
		if err != nil {
			install = true
			log.Println("Database is not found")
		}

		if install == true {
			_, err := os.Create(TODO_DBFILE)
			if err != nil {
				log.Println(err)
			}
			db, err := sql.Open("sqlite", TODO_DBFILE)
			if err != nil {
				log.Println(err)
			}
			defer db.Close()
			err = createTable(db)
			if err != nil {
				log.Println(err)
			}
			fmt.Println("Database is installed with TODO_DBFILE path")
		}
	}
	if exists == false || TODO_DBFILE == "" {
		_, err := os.Stat(defaultDbPath)
		if err != nil {
			install = true
		}
		if install == true {
			_, err := os.Create(defaultDbPath)
			if err != nil {
				log.Println(err)
			}
			db, err := sql.Open("sqlite", defaultDbPath)
			if err != nil {
				log.Println(err)
			}
			defer db.Close()
			err = createTable(db)
			if err != nil {
				log.Println(err)
			}
			fmt.Println("Database is installed with default path")
		}
	}
}
