package main

import "path/filepath"

// Flags list
var (
	port                = 4000
	storagePath         = "data"
	bucketsMetadataPath = filepath.Join(storagePath, "buckets.csv")
)
