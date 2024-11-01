package web

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Initialization loading buckets
func Init() error {
	// Parse flags
	err := Parse(os.Args[1:])
	if err != nil {
		return err
	}
	// Initialize Buckets map
	if bucketMap == nil {
		bucketMap = make(map[string]*bucketData)
	}

	err = loadBucketsData()
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

// Load buckets from data path if exist
// if doesn't exist, creates new buckets.csv
func loadBucketsData() error {
	// Validate storage path
	for _, prohibitedPath := range ProhibitedStoragePaths {
		if prohibitedPath == strings.Trim(storagePath, "/") {
			return ErrProhibitedStoragePath
		}
	}

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
	bucketsMetadataPath := filepath.Join(storagePath, "buckets.csv")
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
			return ErrInvalidNumberOfFields
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

// Global variable of buckets list
// key string is the name of the bucket
var bucketMap map[string]*bucketData

func saveBucketsData() error {
	// Opening buckets.csv metadata file
	bucketsMetadataPath := filepath.Join(storagePath, "buckets.csv")
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

func saveObjectsData(bucketName, objectName string) error {
	bucketMetadataPath := filepath.Join(storagePath, bucketName, "objects.csv")
	bucketMetadataFile, err := os.Create(bucketMetadataPath)
	if err != nil {
		return fmt.Errorf("error while reading <%s> bucket metadata file: %w", bucketName, err)
	}

	csvWriter := csv.NewWriter(bucketMetadataFile)
	for _, object := range *bucketMap[bucketName].objects {
		csvWriter.Write([]string{object.objectKey, strconv.Itoa(object.contentLength), object.contentType, object.lastModified})
	}

	csvWriter.Flush()
	err = csvWriter.Error()
	if err != nil {
		return fmt.Errorf("error while writing metadata to <%s> file: %w", bucketMetadataPath, err)
	}

	// Update metadata in buckets.csv file
	if len(*bucketMap[bucketName].objects) == 0 {
		bucketMap[bucketName].Status = "inactive"
	} else {
		bucketMap[bucketName].Status = "active"
	}
	bucketMap[bucketName].LastModifiedTime = time.Now().Format(time.RFC822)
	// Sync with disk files
	err = saveBucketsData()
	if err != nil {
		return fmt.Errorf("error while saving buckets in <%s> object and <%s> bucket metadata: %w", objectName, bucketName, err)
	}

	return nil
}
