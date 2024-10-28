package main

import "net/http"

func retrieveObject(w http.ResponseWriter, r *http.Request, bucketName, objectName string) {
	w.Write([]byte("Returning the " + objectName + " from " + bucketName))
}

func createObject(w http.ResponseWriter, r *http.Request, bucketName, objectName string) {
	w.Write([]byte("Creating the object " + bucketName + " " + objectName))
}

func deleteObject(w http.ResponseWriter, r *http.Request, bucketName, objectName string) {
	w.Write([]byte("Deleting the object from " + bucketName + " object: " + objectName))
}
