package web

import (
	"encoding/csv"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Errors
var (
	ErrInvalidNumberOfFields = errors.New("incorrect number of fields in csv file")
	ErrBucketAlreadyExists   = errors.New("such bucket already exists")
	ErrBucketNotExists       = errors.New("the specified bucket does not exist")
	ErrBucketContainsDir     = errors.New("bucket contains directory")
	ErrBucketIsNotEmpty      = errors.New("bucket is not empty")

	ErrProhibitedStoragePath = errors.New("prohibited storage path used")
	ErrProhibitedBucketName  = errors.New("bucket name is prohibited")
)

// bucketData csv record structure
type bucketData struct {
	Name             string `xml:"Name"`
	CreatedTime      string `xml:"CreationDate"`
	LastModifiedTime string `xml:"LastModifiedDate"`
	Status           string `xml:"Status"`
	objects          *[]bucketObject
}

var ProhibitedStoragePaths = []string{
	"cmd",
	"scripts",
}

type bucketsWrapper struct {
	XMLName xml.Name      `xml:"Buckets"`
	Buckets []*bucketData `xml:"Bucket"`
}

// GET handler
func getBuckets(w http.ResponseWriter) error {
	wrapper := bucketsWrapper{}

	for _, bucket := range bucketMap {
		wrapper.Buckets = append(wrapper.Buckets, bucket)
	}
	marshalledObject, err := xml.MarshalIndent(wrapper, "", "    ")
	if err != nil {
		return fmt.Errorf("error while marshaling the buckets: %w", err)
	}
	respondSuccessXML(w, marshalledObject)
	log.Print("Buckets list requested")
	return nil
}

var prohibitedBucketNames = []string{
	"buckets.csv",
}

// PUT handler
func createBucket(bucketName string) error {
	// Validate bucket name
	for _, prohibitedName := range prohibitedBucketNames {
		if prohibitedName == bucketName {
			return ErrProhibitedBucketName
		}
	}
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
	} else {
		log.Print("bucket directory created: " + filepath.Join(storagePath, bucketName))
	}

	// Create bucket metadata file
	objectsMetadataPath := filepath.Join(storagePath, bucketName, "objects.csv")
	_, err = os.Create(objectsMetadataPath)
	if err != nil {
		if os.IsExist(err) {
			log.Print("bucket metadata file already exists: " + objectsMetadataPath)
		} else {
			log.Fatal(err)
		}
	} else {
		log.Print("bucket metadata file created: " + objectsMetadataPath)
	}

	// Bucket add to map
	if len(bucketMap) == 0 {
		bucketMap = make(map[string]*bucketData)
	}

	bucketMap[bucketName] = &bucketData{
		Name:             bucketName,
		CreatedTime:      time.Now().Format(time.RFC822),
		LastModifiedTime: time.Now().Format(time.RFC822),
		Status:           "inactive",
		objects:          &[]bucketObject{},
	}

	// write to csv file
	bucketsMetadataPath := filepath.Join(storagePath, "buckets.csv")
	csvFile, err := os.OpenFile(bucketsMetadataPath, os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer csvFile.Close()

	// write new bucket to csv file
	csvWriter := csv.NewWriter(csvFile)
	csvWriter.Write([]string{bucketName, bucketMap[bucketName].CreatedTime, bucketMap[bucketName].LastModifiedTime, bucketMap[bucketName].Status})
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
	bucketDir, err := os.ReadDir(bucketPath)
	if err != nil {
		return fmt.Errorf("error while reading the <%s>  directory: %w", bucketPath, err)
	}
	if len(bucketDir) <= 1 {
		if len(bucketDir) == 1 {
			if bucketDir[0].Name() == "objects.csv" {
				// Remove the bucket's metadata
				bucketMetadataPath := filepath.Join(storagePath, bucketName, "objects.csv")
				err = os.Remove(bucketMetadataPath)
				if err != nil {
					return fmt.Errorf("error while deleting metadata file in <%s> bucket: %w", bucketName, err)
				}
				err = os.Remove(bucketPath)
				if err != nil {
					return fmt.Errorf("error while removing <%s> bucket directory: %w", bucketName, err)
				}
			} else {
				return ErrBucketIsNotEmpty
			}
		}
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

	// Remove the bucket's metadata
	bucketMetadataPath := filepath.Join(storagePath, bucketName, "objects.csv")
	err = os.Remove(bucketMetadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		} else {
			return err
		}
	}

	// Delete the bucket from the map
	delete(bucketMap, bucketName)

	err = saveBucketsData()
	if err != nil {
		return fmt.Errorf("error while saving buckets metadata: %w", err)
	}

	log.Print("<" + bucketName + "> bucket deleted")
	return nil
}
