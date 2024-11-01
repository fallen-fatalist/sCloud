package main

import (
	"fmt"
	"log"
	"net/http"

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

	mux := web.Routes()

	log.Print("starting server on: ", web.Port)

	err = http.ListenAndServe(fmt.Sprintf(":%d", web.Port), mux)

	log.Fatal(err)
}
