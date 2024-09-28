// util.go
// Package main provides utility functions for basic operations such as calculating
// the minimum and maximum of two integers, and retrieving the size of a file.

package main

import (
	"log"
	"os"
)

// min returns the smaller of two integers.
// If both integers are equal, it returns the first one.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the larger of two integers.
// If both integers are equal, it returns the first one.
func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

// getFileSize returns the size of the file specified by filePath in bytes.
// It logs a fatal error and terminates the program if the file cannot be accessed.
func getFileSize(filePath string) int64 {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Fatalf("Error accessing file '%s': %v", filePath, err)
	}
	return fileInfo.Size()
}
