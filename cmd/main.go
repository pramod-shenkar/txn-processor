package main

import (
	"txn-processor/server"
)

func main() {
	app := server.New()
	app.Start()
}
