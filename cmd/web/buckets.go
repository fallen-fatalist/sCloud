package main

import (
	"encoding/csv"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Errors
var (
	ErrIncorrectNumberOfFields = errors.New("incorrect number of fields in csv file")
	ErrBucketAlreadyExists     = errors.New("such bucket already exists")
	ErrBucketNotExists         = errors.New("such bucket  does not exist")
	ErrBucketContainsDir       = errors.New("bucket contains directory")
	ErrBucketIsNotEmpty        = errors.New("bucket is not empty")
)

// Global variable of buckets list
// key string is the name of the bucket
var bucketMap map[string]*bucketData

// bucketData csv record structure
type bucketData struct {
	Name             string `xml:"Name"`
	CreatedTime      string `xml:"CreationDate"`
	LastModifiedTime string `xml:"LastModifiedDate"`
	Status           string `xml:"Status"`
	objects          *[]bucketObject
}

// Load buckets from data path if exist
// if doesn't exist, creates new buckets.csv
func loadBucketsData() error {
	// Create storage directory if not exists
	err := os.Mkdir(storagePath, 0o755)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("error while creating storage directory: %w", err)
		}
	} else {
		log.Print("created storage directory: " + storagePath)
	}

	// Opening buckets.csv metadata file
	bucketsFile, err := os.OpenFile(bucketsMetadataPath, os.O_RDONLY, 0o644)
	if err != nil {
		if os.IsNotExist(err) {
			_, err := os.OpenFile(bucketsMetadataPath, os.O_CREATE, 0o644)
			if err != nil {
				log.Fatal(err)
			} else {
				log.Print("created buckets.csv metadata file")
				return nil
			}
		}
		return fmt.Errorf("error while opening bucket metadata file: %w", err)
	}
	defer bucketsFile.Close()

	// Validation must be in router, so here it is skipped
	// so here bucketName is not validated

	// Initialize Buckets map
	if bucketMap == nil {
		bucketMap = make(map[string]*bucketData)
	}

	// Parsing buckets.csv file
	bucketsCsvReader := csv.NewReader(bucketsFile)

	// Iterate over csv records
	for {
		bucketsRecord, err := bucketsCsvReader.Read()
		if err != nil {
			if err == io.EOF {
				log.Print("loaded buckets metadata")
				return nil
			}
			return fmt.Errorf("error while reading buckets' metadata: %w", err)
			// csv record length validation
		} else if len(bucketsRecord) != 4 {
			return ErrIncorrectNumberOfFields
			// duplication detection
		} else if _, exists := bucketMap[bucketsRecord[0]]; exists {
			return ErrBucketAlreadyExists
		}

		// Object's metadata reading
		objects := []bucketObject{}
		bucketPath := filepath.Join(storagePath, bucketsRecord[0])
		objectMetaDataPath := filepath.Join(bucketPath, "objects.csv")
		objectMetaDataFile, err := os.OpenFile(objectMetaDataPath, os.O_RDONLY, 0o644)
		if err != nil {
			if os.IsNotExist(err) {
				_, err = os.OpenFile(objectMetaDataPath, os.O_CREATE, 0o644)
				if err != nil {
					return fmt.Errorf("error while reading the <%s> bucket and creating its metadata file: %w", bucketsRecord[0], err)
				}
				continue
			}
			return fmt.Errorf("error while reading <%s> object metadata file: %w", bucketsRecord[0], err)
		}
		bucketCsvReader := csv.NewReader(objectMetaDataFile)
		for {
			bucketRecord, err := bucketCsvReader.Read()
			if err != nil {
				if err == io.EOF {
					break
				} else {
					return fmt.Errorf("error while reading <%s> bucket's metadata file: %w", bucketsRecord[0], err)
				}
			}

			length, err := strconv.Atoi(bucketRecord[1])
			if err != nil {
				return fmt.Errorf("error while converting <%s> object length to integer: %w", bucketRecord[0], err)
			}
			objects = append(objects, bucketObject{
				objectKey:     bucketRecord[0],
				contentLength: length,
				contentType:   bucketRecord[2],
				lastModified:  bucketRecord[3],
			})
		}
		objectMetaDataFile.Close()

		// add bucket to bucket map
		bucketMap[bucketsRecord[0]] = &bucketData{
			Name:             bucketsRecord[0],
			CreatedTime:      bucketsRecord[1],
			LastModifiedTime: bucketsRecord[2],
			Status:           bucketsRecord[3],
			objects:          &objects,
		}
	}
}

func saveBucketsData() error {
	// Opening buckets.csv metadata file
	bucketsFile, err := os.Create(bucketsMetadataPath)
	if err != nil {
		return fmt.Errorf("error while opening bucket metadata file: %w", err)
	}
	defer bucketsFile.Close()

	csvWriter := csv.NewWriter(bucketsFile)
	for bucketName, bucketData := range bucketMap {
		err := csvWriter.Write([]string{bucketName, bucketData.CreatedTime, bucketData.LastModifiedTime, bucketData.Status})
		if err != nil {
			return fmt.Errorf("error while saving bucket's metadata to buckets.csv file")
		}
	}
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("error while saving bucket's metadata to buckets.csv file: %w", err)
	}

	return nil
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
	bucketMap[bucketName] = &bucketData{
		Name:             bucketName,
		CreatedTime:      time.Now().Format(time.RFC822),
		LastModifiedTime: time.Now().Format(time.RFC822),
		Status:           "inactive",
		objects:          &[]bucketObject{},
	}

	// write to csv file
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
