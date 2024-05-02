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
		    date VARCHAR(8) ,
		    title TEXT  ,
		    comment TEXT  ,
		    repeat VARCHAR(128) 
		    );
		CREATE INDEX scheduler_date ON scheduler(date);
		`)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

// DatabaseCheck проверяет наличие базы данных, в случае отсутствия создает ее вместе с таблицей scheduler
func DatabaseCheck() {
	var dbPath string
	TODO_DBFILE := os.Getenv("TODO_DBFILE")
	if TODO_DBFILE != "" {
		dbPath = TODO_DBFILE
	} else {
		dbPath = "./scheduler.db"
	}
	var install bool
	// При наличии файла .env c заданным значением переменной TODO_DBFILE база данных будет создана по указанному пути, если переменная не задана, то будет создана в папке database
	_, err := os.Stat(dbPath)
	if err != nil {
		install = true
		log.Println("База данных не установлена", err)
	}

	if install == true {
		_, err := os.Create(dbPath)
		if err != nil {
			log.Println("Ошибка создания базы данных", err)
		}
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			log.Println("Ошибка открытия базы данных", err)
		}
		defer db.Close()
		err = createTable(db)
		if err != nil {
			log.Println(err)
		}
		fmt.Println("База данных установлена")
	}
}
