package main

import (
	"main/api"
)

func main() {
	api.EnvInit()
	api.DatabaseCheck()
	api.StartServer()

}
