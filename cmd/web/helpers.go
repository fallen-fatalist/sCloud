package web

import (
	"encoding/xml"
	"log"
	"net/http"
)

const (
	// Declaration is a generic XML header suitable for use with the output of [Marshal].
	// This is not automatically added to any output of this package,
	// it is provided as a convenience.
	Declaration = `<?xml version="1.0" encoding="UTF-8"?>` + "\n"
)

func respondSuccessXML(w http.ResponseWriter, marshalledObject []byte) {
	w.Header().Set("Content-Type", "application/xml")
	marshalledObject = append(marshalledObject, '\n')
	w.Write(append([]byte(Declaration), marshalledObject...))
}

type errorWrapper struct {
	XMLName  xml.Name `xml:"Error"`
	Code     string   `xml:"Code"`
	Message  string   `xml:"Message"`
	Resource string   `xml:"Resource"`
}

func respondError(w http.ResponseWriter, r *http.Request, statusCode int, err error) {
	message, code := mapErrorToMessageAndCode(err)
	xmlError := errorWrapper{
		Message:  message,
		Code:     code,
		Resource: r.URL.String(),
	}
	marshalledError, innerErr := xml.MarshalIndent(xmlError, "", "    ")
	if innerErr != nil {
		log.Print("error while marshaling the error: %w", innerErr)
	} else {
		marshalledError = append(marshalledError, '\n')
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(statusCode)
		w.Write(append([]byte(Declaration), marshalledError...))
		log.Printf("error while executing URL: %s; with error: %s", r.URL.String(), err)
	}
}

// Error messages
const (
	ResourceDoesNotExist        = "The resource you requested does not exist"
	ResourceAlreadyExists       = "The resource you requested already exists"
	ResourceIdentifierIncorrect = "The resource identifier you requested is incorrect"
	ErrEntityTooLarge           = "Your proposed upload exceeds the maximum allowed object size"
	RequestIncompleteBody       = "You did not provide the number of bytes specified by the Content-Length HTTP header"
	ErrInvalidArgument          = "The resource identifier is incorrect"
	ErrNoSuchKey                = "The specified key does not exist"
)

// Error codes
const (
	IncompleteBody           = "IncompleteBody"
	MaxMessageLengthExceeded = "MaxMessageLengthExceeded"
	NoSuchResource           = "NoSuchResource"
	InvalidArgument          = "InvalidArgument"
	EntityTooLarge           = "EntityTooLarge"
	BucketAlreadyExists      = "BucketAlreadyExists"
	NoSuchBucket             = "NoSuchBucket"
	NoSuchKey                = "NoSuchKey"
	MethodNotAllowed         = "MethodNotAllowed"
	BadRequest               = "BadReqest"
	ExistingKey              = "ExistingKey"
)

// Map certain error to general message message, code is more certain
func mapErrorToMessageAndCode(err error) (message string, code string) {
	// General error messages
	switch err {
	case ErrBucketNotExists:
		message, code = ErrBucketNotExists.Error(), NoSuchBucket
	case ErrBucketAlreadyExists:
		message, code = ErrBucketAlreadyExists.Error(), BucketAlreadyExists
	case ErrObjectNotExists:
		message, code = ErrNoSuchKey, NoSuchKey
	case ErrObjectAlreadyExists:
		message, code = ErrObjectAlreadyExists.Error(), ExistingKey
	case ErrConsecutiveHyphenDot,
		ErrInvalidCharacters,
		ErrManySegments,
		ErrStartWithHyphen,
		ErrTooLongName,
		ErrTooShortName,
		ErrValidIPAddress,
		ErrEndWithHyphenDot:

		message, code = err.Error(), InvalidArgument
	case ErrTooBigObject:
		message, code = ErrEntityTooLarge, MaxMessageLengthExceeded
	case ErrUndefinedContentLength:
		message, code = RequestIncompleteBody, IncompleteBody
	case ErrNoSuchResource:
		message, code = ErrNoSuchResource.Error(), NoSuchResource
	case ErrMethodNotAllowed:
		message, code = ErrMethodNotAllowed.Error(), MethodNotAllowed
	default:
		message, code = err.Error(), BadRequest
	}

	return
}
