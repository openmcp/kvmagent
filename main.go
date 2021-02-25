package main

import (
	"log"
	"net/http"
)

func main() {

	router := NewRouter()
	log.Println("Server Start")
	log.Fatal(http.ListenAndServe(":10000", router))
}
