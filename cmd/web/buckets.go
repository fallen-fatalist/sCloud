package main

import (
	"encoding/csv"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Errors
var (
	ErrIncorrectNumberOfFields = errors.New("incorrect number of fields in csv file")
	ErrBucketAlreadyExists     = errors.New("bucket with BucketName already exists")
	ErrBucketNotExists         = errors.New("bucket with BucketName does not exist")
	ErrBucketContainsDir       = errors.New("bucket contains directory")
	ErrBucketIsNotEmpty        = errors.New("bucket is not empty")
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
	// Create storage directory if not exists
	err := os.Mkdir(storagePath, 0o755)
	if err != nil && !os.IsExist(err) {
		return err
	} else if err == nil {
		log.Print("created storage directory: " + storagePath)
	}

	// Opening buckets.csv metadata file
	bucketsFile, err := os.OpenFile(bucketsPath, os.O_RDONLY, 0o644)
	if err != nil {
		if os.IsNotExist(err) {
			_, err := os.OpenFile(bucketsPath, os.O_CREATE, 0o644)
			if err != nil {
				log.Fatal(err)
			}
			log.Print("created buckets.csv metadata file")
			return nil
		}
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

	// Iterate over csv records
	for {
		record, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				log.Print("loaded buckets metadata")
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
func getBuckets(w http.ResponseWriter) {
	for bucketName, bucketData := range bucketMap {
		w.Write([]byte(bucketName + " " + bucketData.createdTime + " " + bucketData.lastModifiedTime + " " + bucketData.status + "\n"))
	}
	log.Print("Buckets list requested")
}

// PUT handler
func createBucket(bucketName string) error {
	// Name existence check
	if _, exists := bucketMap[bucketName]; exists {
		return ErrBucketAlreadyExists
	}

	// Create bucket directory
	err := os.Mkdir(filepath.Join(storagePath, bucketName), 0o755)
	if err != nil {
		if os.IsExist(err) {
			log.Print("bucket directory already exists: " + filepath.Join(storagePath, bucketName))
		} else {
			log.Fatal(err)
		}
	}

	log.Print("bucket directory created: " + filepath.Join(storagePath, bucketName))

	// Bucket add to map
	bucketMap[bucketName] = BucketInfo{
		createdTime:      time.Now().Format(time.RFC822),
		lastModifiedTime: time.Now().Format(time.RFC822),
		status:           "active",
	}

	// write to csv file
	csvFile, err := os.OpenFile(bucketsPath, os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer csvFile.Close()

	// write new bucket to csv file
	csvWriter := csv.NewWriter(csvFile)
	csvWriter.Write([]string{bucketName, bucketMap[bucketName].createdTime, bucketMap[bucketName].lastModifiedTime, bucketMap[bucketName].status})
	csvWriter.Flush()
	err = csvWriter.Error()
	if err != nil {
		return err
	}

	log.Print(bucketName + " empty bucket created")
	return nil
}

// DELETE handler
func deleteBucket(bucketName string) (err error) {
	if _, exists := bucketMap[bucketName]; !exists {
		return ErrBucketNotExists
	}

	// Remove the bucket's directory
	bucketPath := filepath.Join(storagePath, bucketName)
	bucketDir, _ := os.ReadDir(bucketPath)
	if len(bucketDir) == 0 {
		err = os.Remove(bucketPath)
	} else {
		return ErrBucketIsNotEmpty
	}
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		} else {
			return err
		}
	}

	// Delete the bucket from the map
	delete(bucketMap, bucketName)

	// Overwrite the map to the csv file
	csvFile, err := os.Create(bucketsPath)
	if err != nil {
		return err
	}

	// Write to the new file
	csvWriter := csv.NewWriter(csvFile)
	for bucketName, bucketData := range bucketMap {
		csvWriter.Write([]string{bucketName, bucketData.createdTime, bucketData.lastModifiedTime, bucketData.status})
	}
	csvWriter.Flush()
	err = csvWriter.Error()
	if err != nil {
		return err
	}
	log.Print("<" + bucketName + "> bucket deleted")
	return nil
}
