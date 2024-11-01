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
	ErrNoSuchResource       = errors.New("the specified resource doesn't exist")
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
	log.Printf("%s request with URL: %s", r.Method, r.URL.String())

	// Routing
	switch {
	// / Index route processing
	case r.URL.String() == "/":
		if r.Method == http.MethodGet {
			getBuckets(w)
			return
		} else {
			w.Header().Set("Allow", "GET")
			respondError(w, r, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
			return
		}
	// /<BucketName> route processing
	case len(URLSegments) == 1:
		// URL validation
		err := validateURLSegments(URLSegments)
		if err != nil {
			respondError(w, r, http.StatusBadRequest, err)
			return
		}

		switch r.Method {
		case http.MethodPut:
			err := createBucket(URLSegments[0])
			if err != nil {
				statusCode := http.StatusBadRequest
				if err == ErrBucketAlreadyExists {
					statusCode = http.StatusConflict
				}
				respondError(w, r, statusCode, err)
				return
			}
			w.Header().Set("Location", "/"+URLSegments[0])
			w.Header().Set("Content-Length", "0")
			w.Header().Set("Connection", "close")
			w.Write([]byte("Created the bucket with name: " + URLSegments[0] + "\n"))
			return
		case http.MethodDelete:
			err := deleteBucket(URLSegments[0])
			if err != nil {
				statusCode := http.StatusBadRequest
				if err == ErrBucketIsNotEmpty {
					statusCode = http.StatusConflict
				} else if err == ErrBucketNotExists {
					statusCode = http.StatusNotFound
				}
				respondError(w, r, statusCode, err)
				return
			} else {
				w.WriteHeader(http.StatusNoContent)
				w.Write([]byte("Deleted the bucket with name: " + URLSegments[0] + "\n"))

			}
			return
		default:
			w.Header().Set("Allow", "PUT, DELETE")
			respondError(w, r, http.StatusMethodNotAllowed, ErrMethodNotAllowed)
			return
		}
	// /<BucketName>/<ObjectName> route
	case len(URLSegments) == 2:
		// URL validation
		err := validateURLSegments(URLSegments)
		if err != nil {
			respondError(w, r, http.StatusBadRequest, err)
			return
		}

		switch r.Method {
		case http.MethodGet:
			err := retrieveObject(w, URLSegments[0], URLSegments[1])
			if err != nil {
				statusCode := 400
				if err == ErrObjectNotExists {
					statusCode = http.StatusNotFound
				}
				respondError(w, r, statusCode, err)
				return

			}
			return
		case http.MethodPut:
			err := uploadObject(r, URLSegments[0], URLSegments[1])
			if err != nil {
				statusCode := 400
				if err == ErrObjectAlreadyExists {
					statusCode = http.StatusConflict
				}
				respondError(w, r, statusCode, err)
				return
			}
			return
		case http.MethodDelete:
			err := deleteObject(URLSegments[0], URLSegments[1])
			if err != nil {
				statusCode := 400
				if err == ErrObjectNotExists {
					statusCode = http.StatusNotFound
				}
				respondError(w, r, statusCode, err)
				return
			} else {
				w.WriteHeader(http.StatusNoContent)
				w.Write([]byte("deleted the object with name: " + URLSegments[1] + " in the bucket " + "<" + URLSegments[0] + ">" + "\n"))
				return
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Header().Set("Allow", "GET, PUT, DELETE")
			w.Write([]byte("Incorrect method entered to operate with objects\n"))
			return
		}
	default:
		respondError(w, r, http.StatusNotFound, ErrNoSuchResource)
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

	// Regex Patterns
	CharactersPattern, err := regexp.Compile("^[a-z0-9.-]+$")
	if err != nil {
		return err
	}

	IPPattern, err := regexp.Compile(`^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)
	if err != nil {
		return err
	}
	StartHyphenDotPattern, err := regexp.Compile(`^[.-]`)
	if err != nil {
		return err
	}
	EndHyphenDotPattern, err := regexp.Compile(`[.-]$`)
	if err != nil {
		return err
	}
	ConsecutiveDotsHyphenPattern, err := regexp.Compile(`(\.\.|--)`)
	if err != nil {
		return err
	}

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
