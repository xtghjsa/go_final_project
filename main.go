package main

import (
	"main/api"
	"main/api/database"
)

func main() {
	// Загрузка конфигурации из файла .env
	api.EnvInit()
	// Проверка на наличие базы данных, создание в случае ее отсутствия
	database.CheckDatabase()
	// Запуск сервера
	defer api.StartServer()
}
