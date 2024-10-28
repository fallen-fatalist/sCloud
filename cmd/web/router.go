package main

import (
	"log"
	"net/http"
	"regexp"
	"strings"
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
	URLString := strings.Trim(r.URL.String(), "/")
	URLSegments := strings.Split(URLString, "/")

	// URL validation
	validateURLSegments(w, r, URLSegments)

	// Routing
	switch {
	// List buckets
	case r.URL.String() == "/":
		if r.Method == http.MethodGet {
			getBuckets(w, r)
			return
		} else {
			w.Header().Set("Allow", "GET")
			http.Error(w, "Incorrect method applied for listing buckets", http.StatusMethodNotAllowed)
			return
		}
	// Buckets operations
	case len(URLSegments) == 1:
		switch r.Method {
		case "PUT":
			createBucket(w, r, URLSegments[0])
			return
		case "DELETE":
			deleteBucket(w, r, URLSegments[0])
		default:
			w.Header().Set("Allow", "PUT, DELETE")
			http.Error(w, "Incorrect method applied to operate with buckets", http.StatusMethodNotAllowed)
			return
		}
	// Objects operations
	case len(URLSegments) == 2:
		switch r.Method {
		case "GET":
			retrieveObject(w, r, URLSegments[0], URLSegments[1])
			return
		case "PUT":
			createObject(w, r, URLSegments[0], URLSegments[1])
		case "DELETE":
			deleteObject(w, r, URLSegments[0], URLSegments[1])
		default:
			w.Header().Set("Allow", "GET, PUT, DELETE")
			http.Error(w, "Incorrect method applied to operate with objects", http.StatusMethodNotAllowed)
			return
		}
	}

}

func validateURLSegments(w http.ResponseWriter, r *http.Request, URLSegments []string) {
	if len(URLSegments) > 2 {
		http.Error(w, "Too many segments in URL string, must be less 3", http.StatusBadRequest)
		return
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
	for idx, segment := range URLSegments {
		segmentBytes := []byte(segment)
		segmentName := ""
		if idx == 0 {
			segmentName = "Bucket"
		} else if idx == 1 {
			segmentName = "Object"
		}

		// length validation
		if len(segment) > 63 {
			incorrectName(w, segmentName, "name too long, must not exceed 63 characters")
			return
		} else if len(segment) < 3 {
			incorrectName(w, segmentName, "name too short, must not be lesser than 3 characters")
			return
		}

		// characters validation
		if match := CharactersPattern.Match(segmentBytes); !match {
			incorrectName(w, segmentName, "name contains not valid characters: only lowercase letters, digits, dots, hyphens allowed")
			return
		}

		// ip validation
		if match := IPPattern.Match(segmentBytes); match {
			incorrectName(w, segmentName, "name matches IP address, name must not match IP addresses")
			return
		}

		// begin with dot/hyphen validation
		if match := StartHyphenDotPattern.Match(segmentBytes); match {
			incorrectName(w, segmentName, "name must not begin with dot or hyphen")
			return
		}
		// end with dot/hyphen validation
		if match := EndHyphenDotPattern.Match(segmentBytes); match {
			incorrectName(w, segmentName, "name must not end with dot or hyphen")
			return
		}
		// consecutive dots/hyphens
		if match := ConsecutiveDotsHyphenPattern.Match(segmentBytes); match {
			incorrectName(w, segmentName, "name must not contain consecutive dots or hyphens")
			return
		}
	}

}

func incorrectName(w http.ResponseWriter, segmentName, message string) {
	http.Error(w, segmentName+" "+message, http.StatusBadRequest)
	return
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
