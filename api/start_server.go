package api

import (
	"database/sql"
	"fmt"
	"log"
	"main/api/database"
	"main/api/handlers"
	"main/api/handlers/authentication"
	"net/http"
	"os"
)

var port = "7540"
var webDir = "./web"

// StartServer Запускает сервер
func StartServer() {
	var TODO_PORT = os.Getenv("TODO_PORT")
	if TODO_PORT != "" {
		port = TODO_PORT
	}
	db, err := sql.Open("sqlite", database.DbPath)
	if err != nil {
		log.Fatal("Ошибка открытия базы данных", err)
	}
	defer db.Close()
	database := &handlers.TaskManager{DB: db}

	savedPass := os.Getenv("TODO_PASSWORD")
	pass := &authentication.Password{SavedPassword: savedPass}

	// При наличии файла .env c заданным значением переменной TODO_PORT сервер запустится на указанном порту если переменная не задана, то будет использован стандартный порт 7540
	fmt.Printf("Сервер запущен на localhost:%s\n", port)
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.HandleFunc("/api/nextdate", handlers.NextDateHandler)
	http.HandleFunc("/api/tasks", pass.Auth(database.ShowTasksHandler))
	http.HandleFunc("/api/task", pass.Auth(database.TaskInteractionHandler))
	http.HandleFunc("/api/task/done", pass.Auth(database.MarkTaskAsDoneHandler))
	http.HandleFunc("/api/signin", pass.CheckPasswordHandler)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
