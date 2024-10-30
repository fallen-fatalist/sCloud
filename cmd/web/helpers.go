package main

import "net/http"

func respondError(statusCode int, err error, w http.ResponseWriter) {
	http.Error(w, err.Error(), statusCode)
}
