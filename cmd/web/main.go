package main

import (
	"fmt"
	"log"
	"net/http"
)

// Config variables
var (
	port = 4000
)

func main() {
	mux := routes()

	log.Print("starting server on:", port)

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux)

	log.Fatal(err)
}
