package main

import (
	"net/http"
)

// GET handler
func getBuckets(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("Returning buckets"))
}

// PUT handler
func createBucket(w http.ResponseWriter, r *http.Request, bucketName string) {
	w.Write([]byte("Created bucket"))
}

// DELETE handlerS
func deleteBucket(w http.ResponseWriter, r *http.Request, bucketName string) {
	w.Write([]byte("Deleted the bucket"))
}
