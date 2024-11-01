package web

import (
	"errors"
	"fmt"
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
	ErrProhibitedObjectName   = errors.New("object's name is prohibited")
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
			objectPath := filepath.Join(storagePath, bucketName, objectName)
			objectFile, err := os.Open(objectPath)
			if err != nil {
				return fmt.Errorf("error while opening <%s> object in <%s> bucket: %w", objectName, bucketName, err)
			}

			// Set length of response
			fileInfo, err := objectFile.Stat()
			if err != nil {
				return fmt.Errorf("error while getting file info of <%s> object in <%s> bucket: %w", objectName, bucketName, err)
			}
			w.Header().Set("Content-Length", strconv.Itoa(int(fileInfo.Size())))

			// Set MimeType of response
			signatureBuf := make([]byte, 512)
			n, err := objectFile.Read(signatureBuf)
			if err != nil && err != io.EOF {
				return fmt.Errorf("error while reading first 512 bytes from <%s> object: %w", objectName, err)
			}
			w.Header().Set("Content-Type", http.DetectContentType(signatureBuf[:n]))
			w.Write(signatureBuf[:n])
			signatureBuf = nil
			buf := make([]byte, 100)
			for {
				n, err := objectFile.Read(buf)
				if err != nil && err != io.EOF {
					return fmt.Errorf("error while reading <%s> object in <%s> bucket: %w", objectName, bucketName, err)
				}
				if n == 0 {
					break
				}
				w.Write(buf)
			}
			buf = nil
			return nil
		}
	}
	return ErrObjectNotExists
}

var prohibitedObjectNames = []string{
	"objects.csv",
}

func uploadObject(r *http.Request, bucketName, objectName string) error {
	// Names validation
	for _, prohibitedName := range prohibitedObjectNames {
		if prohibitedName == objectName {
			return ErrProhibitedObjectName
		}
	}

	// Bucket existence check
	if _, exists := bucketMap[bucketName]; !exists {
		return ErrBucketNotExists
	}

	// Object existence check in the bucket
	for _, object := range *bucketMap[bucketName].objects {
		if object.objectKey == objectName {
			err := deleteObject(bucketName, object.objectKey)
			if err != nil {
				return fmt.Errorf("error while deleting existing object: %w", err)
			}
			break
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
	}
	signatureBuf := make([]byte, 512)
	n, err := r.Body.Read(signatureBuf)
	if err != nil && err != io.EOF {
		return fmt.Errorf("error while reading request body in <%s> object and <%s> bucket: %w", objectName, bucketName, err)
	}
	defer r.Body.Close()

	// Detect the MIME type
	contentType := http.DetectContentType(signatureBuf)

	// Create the file in storage and upload request body into it
	objectPath := filepath.Join(storagePath, bucketName, objectName)
	objectFile, err := os.OpenFile(objectPath, os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return fmt.Errorf("error while creating <%s> object in <%s> bucket: %w", objectName, bucketName, err)
	} else if contentLength > 0 {
		// Write the file
		objectFile.Write(signatureBuf[:n])
		signatureBuf = nil
		buf := make([]byte, 100)
		for {
			n, err := r.Body.Read(buf)
			if err != nil && err != io.EOF {
				return fmt.Errorf("error while reading request body in <%s> object and <%s> bucket: %w", objectName, bucketName, err)
			} else if n == 0 {
				break
			}
			objectFile.Write(buf[:n])

		}
		buf = nil
	}
	defer objectFile.Close()

	// Read the request body sequentially

	// Append metadata
	*bucketMap[bucketName].objects = append(*bucketMap[bucketName].objects, bucketObject{
		objectKey:     objectName,
		contentLength: int(contentLength),
		contentType:   contentType,
		lastModified:  time.Now().Format(time.RFC822),
	})

	err = saveObjectsData(bucketName, objectName)
	if err != nil {
		return fmt.Errorf("error while saving objects metadata in <%s> bucket: %w", bucketName, err)
	}

	return nil
}

func deleteObject(bucketName, objectName string) error {
	// Bucket existence check
	if _, exists := bucketMap[bucketName]; !exists {
		return ErrBucketNotExists
	}

	// Object existence check
	for idx, object := range *bucketMap[bucketName].objects {
		if object.objectKey == objectName {
			// Remove from objects slice
			*bucketMap[bucketName].objects = append((*bucketMap[bucketName].objects)[:idx], (*bucketMap[bucketName].objects)[idx+1:]...)

			// Remove object from bucket in disk
			objectPath := filepath.Join(storagePath, bucketName, objectName)
			err := os.Remove(objectPath)
			if err != nil {
				if os.IsNotExist(err) {
					err = nil
				} else {
					return fmt.Errorf("error while removing <%s> object in <%s> bucket: %w", objectName, bucketName, err)
				}
			}

			// Update metadata in objects.csv file
			err = saveObjectsData(bucketName, objectName)
			if err != nil {
				return fmt.Errorf("error while saving objects metadata in <%s> bucket: %w", bucketName, err)
			}

			return nil
		}
	}
	return ErrObjectNotExists
}
