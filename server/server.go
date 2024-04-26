package server

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

// EnvInit Загружает переменные из файла .env
func EnvInit() {
	if err := godotenv.Load(".env"); err != nil {
		log.Print(".env file notfound")
	}
}

// StartServer Запускает сервер
func StartServer() {
	webDir := "./web"
	// При наличии файла .env c заданным значением переменной TODO_PORT сервер запустится на указанном порту
	TODO_PORT, exists := os.LookupEnv("TODO_PORT")
	if exists {
		fmt.Println("Server is running on localhost:" + TODO_PORT)
		http.Handle("/", http.FileServer(http.Dir(webDir)))
		err := http.ListenAndServe(":"+TODO_PORT, nil)
		if err != nil {
			panic(err)
		}
	}
	// Если переменная не задана, по умолчанию будет использован порт 7540
	fmt.Println("Server is running on localhost:7540")
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	err := http.ListenAndServe(":7540", nil)
	if err != nil {
		panic(err)
	}

}
