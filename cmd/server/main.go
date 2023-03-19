package main

import (
	"gometric/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)

	serv := server.NewServer()
	go serv.ListenAndServe(":8081")
	log.Print("Server started")

	<-sigint

	serv.Shutdown()
	log.Print("Server stopped")
}
