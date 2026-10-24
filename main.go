package main

import (
	"log"
)

func main() {
	store, err := NewPostGresStore()
	if err != nil {
		log.Fatal(err)
	}

	if err := store.Init(); err != nil {
		log.Fatal(err)
	}
	log.Println("[LOG] Successfully connected to DB")

	server := NewAPIServer(":3000", store)
	server.Run()
}
