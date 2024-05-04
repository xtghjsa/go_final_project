package database

import (
	"database/sql"
	"errors"
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
		CREATE INDEX IF NOT EXISTS scheduler_date ON scheduler(date);
		`)
	if err != nil {
		return err
	}
	return nil
}

var DbPath = "./scheduler.db"

// CheckDatabase проверяет наличие базы данных, в случае отсутствия создает ее вместе с таблицей scheduler
func CheckDatabase() {

	TODO_DBFILE := os.Getenv("TODO_DBFILE")
	if TODO_DBFILE != "" {
		DbPath = TODO_DBFILE
	}
	var install bool
	// При наличии файла .env c заданным значением переменной TODO_DBFILE база данных будет создана по указанному пути, если переменная не задана, то будет создана в корне проекта
	_, err := os.Stat(DbPath)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		install = true
		log.Println("База данных не установлена", err)
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatal("Ошибка проверки наличия базы данных", err)
	}
	if install {
		_, err := os.Create(DbPath)
		if err != nil {
			log.Fatal("Ошибка создания базы данных", err)
		}
		db, err := sql.Open("sqlite", DbPath)
		if err != nil {
			log.Fatal("Ошибка открытия базы данных", err)
		}
		defer db.Close()
		err = createTable(db)
		if err != nil {
			log.Fatal("Ошибка создания таблицы", err)
		}
		fmt.Println("База данных установлена")
	}
}
