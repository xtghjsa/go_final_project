package main

import (
	"main/database"
	"main/server"
)

func main() {
	server.EnvInit()
	database.DatabaseCheck()
	server.StartServer()

}
