package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

// Initialization loading buckets
func init() {
	// Parse flags
	err := Parse(os.Args[1:])
	if err != nil {
		if err == ErrHelpCalled {
			os.Exit(0)
		}
	}

	// create
	err = loadBucketsData()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// RESTful API router initialization
	mux := routes()

	log.Print("starting server on: ", port)

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux)

	log.Fatal(err)
}
