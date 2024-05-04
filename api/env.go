package api

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

// EnvInit Загружает переменные из файла .env и создает файл .env в корне проекта если он не существует
func EnvInit() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("файл .env не найден", err)
		log.Println("создаю файл .env")
		_, err = os.Create(".env")
		if err != nil {
			log.Println("не удалось создать файл .env", err)
		}
	}
}
