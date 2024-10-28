package main

import (
	"errors"
	"log"
	"net/http"
	"regexp"
	"strings"
)

// Errors
var (
	ErrInvalidCharacters    = errors.New("object/bucket name contains invalid characters")
	ErrValidIPAddress       = errors.New("object/bucket name is valid IP address, IP address is not allowed")
	ErrTooShortName         = errors.New("object/bucket name is too short, must be longer than 3 characters")
	ErrTooLongName          = errors.New("object/bucket name is too long, must be shorter than 63 characters")
	ErrStartWithHyphen      = errors.New("object/bucket name cannot start with hyphen or dot")
	ErrEndWithHyphenDot     = errors.New("object/bucket name cannot end with hyphen or dot")
	ErrConsecutiveHyphenDot = errors.New("object/bucket name has consecutive hyphens or dots")
	ErrManySegments         = errors.New("too many segments in URL string, must be less 3")
)

func routes() *http.ServeMux {
	mux := http.NewServeMux()
	// routerHandler handles all routes
	mux.HandleFunc("/", routerHandler)

	return mux

}

// Validate the bucket name to ensure it meets Amazon S3 naming requirements (3-63 characters, only lowercase letters, numbers, hyphens, and periods).

// Router handler
func routerHandler(w http.ResponseWriter, r *http.Request) {
	// Dividing URL into segments
	URLStringTrimmed := strings.Trim(r.URL.String(), "/")
	URLSegments := strings.Split(URLStringTrimmed, "/")

	// Routing
	switch {
	// Index
	case r.URL.String() == "/":
		if r.Method == http.MethodGet {
			getBuckets(w, r)
			return
		} else {
			w.Header().Set("Allow", "GET")
			w.Write([]byte("Incorrect method applied for listing buckets"))
			return
		}
	// Buckets operations
	case len(URLSegments) == 1:
		// URL validation
		err := validateURLSegments(URLSegments)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch r.Method {
		case "PUT":
			err := createBucket(w, r, URLSegments[0])
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Write([]byte("Created the bucket"))
			return
		case "DELETE":
			deleteBucket(w, r, URLSegments[0])
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Header().Set("Allow", "PUT, DELETE")
			w.Write([]byte("Incorrect method entered to operate with buckets"))
			return
		}
	// Objects operations
	case len(URLSegments) == 2:
		// URL validation
		err := validateURLSegments(URLSegments)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch r.Method {
		case "GET":
			retrieveObject(w, r, URLSegments[0], URLSegments[1])
			return
		case "PUT":
			createObject(w, r, URLSegments[0], URLSegments[1])
			return
		case "DELETE":
			deleteObject(w, r, URLSegments[0], URLSegments[1])
			return
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Header().Set("Allow", "GET, PUT, DELETE")
			w.Write([]byte("Incorrect method entered to operate with objects"))
			return
		}
	default:
		http.Error(w, "Incorrect URL format entered", http.StatusBadRequest)
		return
	}

}

// Validation of URL variables
func validateURLSegments(URLSegments []string) error {
	if len(URLSegments) > 2 {
		return ErrManySegments
	} else if len(URLSegments) == 0 {
		return nil
	}

	// regex Patterns
	CharactersPattern, err := regexp.Compile("^[a-z0-9.-]+$")
	checkErr(err)

	IPPattern, err := regexp.Compile(`^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
	checkErr(err)
	StartHyphenDotPattern, err := regexp.Compile(`^[.-]`)
	checkErr(err)
	EndHyphenDotPattern, err := regexp.Compile(`[.-]$`)
	checkErr(err)
	ConsecutiveDotsHyphenPattern, err := regexp.Compile(`(\.\.|--)`)
	checkErr(err)

	// bucket and object name validation
	for _, segment := range URLSegments {
		segmentBytes := []byte(segment)

		// length validation
		if len(segment) > 63 {
			return ErrTooLongName
		} else if len(segment) < 3 {
			return ErrTooShortName
		}

		// characters validation
		if match := CharactersPattern.Match(segmentBytes); !match {
			return ErrInvalidCharacters
		}

		// ip validation
		if match := IPPattern.Match(segmentBytes); match {
			return ErrValidIPAddress
		}

		// begin with dot/hyphen validation
		if match := StartHyphenDotPattern.Match(segmentBytes); match {
			return ErrStartWithHyphen
		}
		// end with dot/hyphen validation
		if match := EndHyphenDotPattern.Match(segmentBytes); match {
			return ErrEndWithHyphenDot
		}
		// consecutive dots/hyphens
		if match := ConsecutiveDotsHyphenPattern.Match(segmentBytes); match {
			return ErrConsecutiveHyphenDot
		}
	}

	return nil
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
