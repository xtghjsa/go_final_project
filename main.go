package main

import (
	"main/api"
)

func main() {
	// Загрузка конфигурации из файла .env
	api.EnvInit()
	// Проверка на наличие базы данных, создание в случае ее отсутствия
	api.DatabaseCheck()
	// Запуск сервера
	defer api.StartServer()
}
