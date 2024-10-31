package main

import (
	"fmt"
	"log"
	"net/http"
)

// Config variables
var ()

// Initialization loading buckets
func init() {
	err := loadBucketsData()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// RESTful API router initialization
	mux := routes()

	log.Print("starting server on:", port)

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux)

	log.Fatal(err)
}
