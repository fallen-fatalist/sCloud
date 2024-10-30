package main

import (
	"encoding/csv"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Errors
var (
	ErrObjectNotExists        = errors.New("object does not exist")
	ErrObjectAlreadyExists    = errors.New("object already exists")
	ErrUndefinedContentLength = errors.New("object's size is undefined")
	ErrTooBigObject           = errors.New("object's size is too big")
)

// Number of bytes in 1gb
const bytesIn1gb = 1024 * 1024 * 1024

type bucketObject struct {
	objectKey     string
	contentLength int
	contentType   string
	lastModified  string
}

func retrieveObject(w http.ResponseWriter, bucketName, objectName string) error {
	// Bucket existence check
	if _, exists := bucketMap[bucketName]; !exists {
		return ErrBucketNotExists
	}

	// Object not existence check in the bucket
	for _, object := range *bucketMap[bucketName].objects {
		if object.objectKey == objectName {
			length := strconv.Itoa(object.contentLength)
			w.Write([]byte(object.objectKey + " " + length + " " + object.contentType + " " + object.lastModified))
			return nil
		}
	}
	return ErrObjectNotExists
}

func uploadObject(r *http.Request, bucketName, objectName string) error {
	// Bucket existence check
	if _, exists := bucketMap[bucketName]; !exists {
		return ErrBucketNotExists
	}

	// Object not existence check in the bucket
	for _, object := range *bucketMap[bucketName].objects {
		if object.objectKey == objectName {
			return ErrObjectAlreadyExists
		}
	}

	// Object upload

	// length validation
	contentLength := r.ContentLength
	if contentLength == -1 {
		return ErrUndefinedContentLength
		// 1GB restriction
	} else if contentLength > bytesIn1gb {
		return ErrTooBigObject
	} else if contentLength == 0 {

	}

	// read the entire request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	// Detect the MIME type
	var contentType string
	if r.Header.Get("Content-Type") != "" {
		contentType = r.Header.Get("Content-Type")
	} else {
		contentType = http.DetectContentType(body)
	}

	// Append metadata
	*bucketMap[bucketName].objects = append(*bucketMap[bucketName].objects, bucketObject{
		objectKey:     objectName,
		contentLength: int(contentLength),
		contentType:   contentType,
		lastModified:  time.Now().Format(time.RFC822),
	})

	// update csv file
	metadataPath := filepath.Join(storagePath, bucketName, objectName, "objects.csv")
	csvFile, err := os.OpenFile(metadataPath, os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer csvFile.Close()

	// write new bucket to csv file
	csvWriter := csv.NewWriter(csvFile)
	csvWriter.Write([]string{objectName, strconv.Itoa(int(contentLength)), contentType, time.Now().Format(time.RFC822)})
	csvWriter.Flush()
	err = csvWriter.Error()
	if err != nil {
		return err
	}

	return nil
}

func deleteObject(w http.ResponseWriter, r *http.Request, bucketName, objectName string) {

}
