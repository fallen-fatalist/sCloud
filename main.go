package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/fallen-fatalist/S3Cloud/cmd/web"
)

func main() {
	// RESTful API router initialization
	err := web.Init()
	if err != nil {
		if err == web.ErrHelpCalled {
			return
		}
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
