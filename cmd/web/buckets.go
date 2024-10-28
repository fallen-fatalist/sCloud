package main

import (
	"encoding/csv"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Errors
var (
	ErrIncorrectNumberOfFields = errors.New("incorrect number of fields in csv file")
	ErrBucketAlreadyExists     = errors.New("bucket with BucketName already exists")
)

// Global variable of buckets list
// key string is the name of the bucket
var bucketMap map[string]BucketInfo

// Bucket csv record structure
type BucketInfo struct {
	// name string is included as key in bucketMap
	createdTime      string
	lastModifiedTime string
	status           string
}

// Load buckets from data path if exist
// if doesn't exist, creates new buckets.csv
func loadBuckets() error {
	// Opening file
	bucketsFile, err := os.OpenFile(bucketsPath, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer bucketsFile.Close()
	// Validation must be in router, so here it is skipped
	// so here bucketName is not validated

	// Initialize Buckets map
	if bucketMap == nil {
		bucketMap = make(map[string]BucketInfo)
	}

	// Parsing csv file
	csvReader := csv.NewReader(bucketsFile)

	for {
		record, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
			// csv record length validation
		} else if len(record) != 4 {
			return ErrIncorrectNumberOfFields
			// duplication detection
		} else if _, exists := bucketMap[record[0]]; exists {
			return ErrBucketAlreadyExists
		}

		// add bucket to bucket map
		bucketMap[record[0]] = BucketInfo{
			createdTime:      record[1],
			lastModifiedTime: record[2],
			status:           record[3],
		}
	}

	return nil

}

// GET handler
func getBuckets(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Returning buckets"))
}

// PUT handler
func createBucket(w http.ResponseWriter, r *http.Request, bucketName string) error {
	// Name existence check
	if _, exists := bucketMap[bucketName]; exists {
		return ErrBucketAlreadyExists
	}

	// Bucket add to map
	bucketMap[bucketName] = BucketInfo{
		createdTime:      time.Now().Format(time.RFC822),
		lastModifiedTime: time.Now().Format(time.RFC822),
		status:           "active",
	}

	// write to csv file
	csvFile, err := os.OpenFile(bucketsPath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer csvFile.Close()

	// write new bucket to csv file
	csvWriter := csv.NewWriter(csvFile)
	csvWriter.Write([]string{bucketName, bucketMap[bucketName].createdTime, bucketMap[bucketName].lastModifiedTime, bucketMap[bucketName].status})
	csvWriter.Flush()
	err = csvWriter.Error()
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

// DELETE handler
func deleteBucket(w http.ResponseWriter, r *http.Request, bucketName string) {
	w.Write([]byte("Deleted the bucket"))
}
