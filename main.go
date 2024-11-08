package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/fallen-fatalist/triple-s/cmd/web"
)

func main() {
	// RESTful API router initialization
	err := web.Init()
	if err != nil {
		if err == web.ErrHelpCalled {
			return
		}
		log.Fatalf("error while initialization application: %s", err)
	}
	if web.Port == 0 {
		log.Print("0 port prohibited")
		os.Exit(1)
	}

	mux := web.Routes()

	log.Print("starting server on: ", web.Port)

	err = http.ListenAndServe(fmt.Sprintf(":%d", web.Port), mux)

	log.Fatal(err)
}
