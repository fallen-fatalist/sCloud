package main

import "path/filepath"

// Flags list
var (
	port        = 4000
	storagePath    = "data"
	bucketsPath = filepath.Join(storagePath, "buckets.csv")
)
