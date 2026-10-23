package main

import (
	"fmt"
	"log"
)

func main() {
	store, err := NewPostGresStore()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("%+v\n", store)

	// server := NewAPIServer(":3000")
	// server.Run()
}
