package api

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"time"
)

// EnvInit Загружает переменные из файла .env
func EnvInit() {
	if err := godotenv.Load(".env"); err != nil {
		log.Print(".env file notfound")
	}
}

// StartServer Запускает сервер
func StartServer() {
	defaultPort := "7540"
	webDir := "./web"
	// При наличии файла .env c заданным значением переменной TODO_PORT сервер запустится на указанном порту
	TODO_PORT, exists := os.LookupEnv("TODO_PORT")
	if exists && TODO_PORT != "" {
		fmt.Printf("Server is running on localhost:%s", TODO_PORT+" (from .env)")
		http.Handle("/", http.FileServer(http.Dir(webDir)))
		http.HandleFunc("/api/nextdate", NextDateHandler)
		err := http.ListenAndServe(":"+TODO_PORT, nil)
		if err != nil {
			panic(err)
		}
	}
	// Если переменная не задана, по умолчанию будет использован стандартный порт 7540
	fmt.Printf("Server is running on localhost:%s", defaultPort+" (default)")
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	err := http.ListenAndServe(":7540", nil)
	if err != nil {
		panic(err)
	}

}
func NextDateHandler(w http.ResponseWriter, r *http.Request) {
	nowParameter := r.URL.Query().Get("now")
	dateParameter := r.URL.Query().Get("date")
	repeatParameter := r.URL.Query().Get("repeat")
	now, err := time.Parse("20060102", nowParameter)
	if err != nil {
		http.Error(w, "incorrect now format", http.StatusBadRequest)
		return
	}
	next, err := NextDate(now, dateParameter, repeatParameter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err = fmt.Fprintf(w, next)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
