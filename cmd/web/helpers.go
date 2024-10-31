package main

import (
	"log"
	"net/http"
)

const (
	// Declaration is a generic XML header suitable for use with the output of [Marshal].
	// This is not automatically added to any output of this package,
	// it is provided as a convenience.
	Declaration = `<?xml version="1.0" encoding="UTF-8"?>` + "\n"
)

func respondSuccessXML(w http.ResponseWriter, marshalledObject []byte) {
	w.Header().Set("Content-Type", "application/xml")
	marshalledObject = append(marshalledObject, '\n')
	w.Write(append([]byte(Declaration), marshalledObject...))
}

func respondError(w http.ResponseWriter, r *http.Request, statusCode int, err error) {
	http.Error(w, err.Error(), statusCode)
	log.Printf("error while executing URL: %s; with error:%s", r.URL.String(), err)
}
