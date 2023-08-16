package main

import (
	"internal/server"
)

func main() {
	serv := server.NewServer()
	serv.ListenAndServe(":8080")
}
