// values_test.go
// Package main contains tests for utility functions related to value matching and conversion.
// These tests verify the correctness of functions like getLongestMatchPosAndLen, getMatchIndex,
// BytesToValues, and ValuesToBytes.

package main

import (
	"crypto/rand"
	"fmt"
	"testing"
)

// Test_getLongestMatchPosAndLen tests the getLongestMatchPosAndLen function.
// It verifies that the function correctly identifies the longest matching substring
// between the search buffer and the lookahead buffer based on the minimum match length.
func Test_getLongestMatchPosAndLen(t *testing.T) {
	tests := []struct {
		name          string
		searchBuff    []byte
		lookaheadBuff []byte
		wantPos       int
		wantLen       byte
		minMatchLen   byte
	}{
		{
			name:          "Empty search buffer",
			searchBuff:    []byte(""),
			lookaheadBuff: []byte("hijklmno"),
			wantPos:       0,
			wantLen:       0,
			minMatchLen:   0,
		},
		{
			name:          "Empty lookahead buffer",
			searchBuff:    []byte("abcdefg"),
			lookaheadBuff: []byte(""),
			wantPos:       0,
			wantLen:       0,
			minMatchLen:   0,
		},
		{
			name:          "No matches in lookahead buffer",
			searchBuff:    []byte("abcdefg"),
			lookaheadBuff: []byte("hijklmno"),
			wantPos:       0,
			wantLen:       0,
			minMatchLen:   0,
		},
		{
			name:          "Full match",
			searchBuff:    []byte("abcdefg"),
			lookaheadBuff: []byte("abcdefg"),
			wantPos:       0,
			wantLen:       7,
			minMatchLen:   0,
		},
		{
			name:          "Half match in search buffer",
			searchBuff:    []byte("abc"),
			lookaheadBuff: []byte("abcdefg"),
			wantPos:       0,
			wantLen:       3,
			minMatchLen:   0,
		},
		{
			name:          "Half match in lookahead buffer",
			searchBuff:    []byte("abcdefg"),
			lookaheadBuff: []byte("abc"),
			wantPos:       0,
			wantLen:       3,
			minMatchLen:   0,
		},
		{
			name:          "Second half of search buffer matches",
			searchBuff:    []byte("efgabc"),
			lookaheadBuff: []byte("abc"),
			wantPos:       3,
			wantLen:       3,
			minMatchLen:   0,
		},
		{
			name:          "Two matches, first one is longer",
			searchBuff:    []byte("milk milk"),
			lookaheadBuff: []byte("milk "),
			wantPos:       0,
			wantLen:       5,
			minMatchLen:   0,
		},
		{
			name:          "Full match shorter than minMatchLen",
			searchBuff:    []byte("abcdefgh"),
			lookaheadBuff: []byte("abcdefgh "),
			wantPos:       0,
			wantLen:       0,
			minMatchLen:   9,
		},
		{
			name:          "Random match in the middle",
			searchBuff:    []byte("abcd peace efgh"),
			lookaheadBuff: []byte(" peace abcd "),
			wantPos:       4,
			wantLen:       7,
			minMatchLen:   5,
		},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Run tests in parallel for efficiency

			pos, length := getLongestMatchPosAndLen(tt.searchBuff, tt.lookaheadBuff, tt.minMatchLen)

			if pos != tt.wantPos {
				t.Errorf("getLongestMatchPosAndLen() pos = %d; want %d", pos, tt.wantPos)
			}
			if length != tt.wantLen {
				t.Errorf("getLongestMatchPosAndLen() len = %d; want %d", length, tt.wantLen)
			}
			if length != 0 && length < tt.minMatchLen {
				t.Errorf("getLongestMatchPosAndLen() found match shorter than minMatchLen: got len %d, min len %d", length, tt.minMatchLen)
			}
		})
	}
}

// Test_getMatchIndex tests the getMatchIndex function.
// It verifies that the function correctly identifies all starting indices
// where the pattern occurs within the text.
func Test_getMatchIndex(t *testing.T) {
	tests := []struct {
		name        string
		text        []byte
		pattern     []byte
		wantMatches []int
	}{
		{
			name:        "Match starts at the beginning",
			text:        []byte("hello"),
			pattern:     []byte("hel"),
			wantMatches: []int{0},
		},
		{
			name:        "Match starts somewhere in the middle",
			text:        []byte("abchello"),
			pattern:     []byte("hel"),
			wantMatches: []int{3},
		},
		{
			name:        "Pattern is empty",
			text:        []byte("abchello"),
			pattern:     []byte(""),
			wantMatches: []int{0, 1, 2, 3, 4, 5, 6, 7},
		},
		{
			name:        "Text is empty",
			text:        []byte(""),
			pattern:     []byte("abc"),
			wantMatches: []int{},
		},
		{
			name:        "Several matches",
			text:        []byte("aaaabcaaaabcaaaabc"),
			pattern:     []byte("abc"),
			wantMatches: []int{4, 10, 16},
		},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Run tests in parallel for efficiency

			gotMatches := getMatchIndex(tt.text, tt.pattern)

			if len(gotMatches) != len(tt.wantMatches) {
				t.Errorf("getMatchIndex() returned %v; want %v", gotMatches, tt.wantMatches)
				return
			}

			for i, match := range gotMatches {
				if match != tt.wantMatches[i] {
					t.Errorf("getMatchIndex()[%d] = %d; want %d", i, match, tt.wantMatches[i])
				}
			}
		})
	}
}

// Test_bytesToValues tests the BytesToValues function.
// It verifies that the function correctly converts byte slices into a series of Value representations
// based on matching criteria such as minimum and maximum match lengths and search buffer length.
func Test_bytesToValues(t *testing.T) {
	tests := []struct {
		name             string
		input            []byte
		minMatchLen      byte
		maxMatchLen      byte
		maxSearchBuffLen uint16
		wantValuesRepr   string
	}{
		{
			name:             "No matches at all",
			input:            []byte("abcd"),
			minMatchLen:      0,
			maxMatchLen:      255,
			maxSearchBuffLen: 255,
			wantValuesRepr:   "abcd",
		},
		{
			name:             "Match at the end",
			input:            []byte("abcd abcd"),
			minMatchLen:      0,
			maxMatchLen:      255,
			maxSearchBuffLen: 255,
			wantValuesRepr:   "abcd <5,4>",
		},
		{
			name:             "Match in the middle",
			input:            []byte("abcd abcd ghij"),
			minMatchLen:      0,
			maxMatchLen:      255,
			maxSearchBuffLen: 255,
			wantValuesRepr:   "abcd <5,5>ghij",
		},
		{
			name:             "Two matches with same length",
			input:            []byte("XXabXXcdXX"),
			minMatchLen:      2,
			maxMatchLen:      255,
			maxSearchBuffLen: 255,
			wantValuesRepr:   "XXab<4,2>cd<8,2>",
		},
		{
			name:             "Three matches with same length",
			input:            []byte("XXabXXcdXXijXX"),
			minMatchLen:      2,
			maxMatchLen:      255,
			maxSearchBuffLen: 255,
			wantValuesRepr:   "XXab<4,2>cd<8,2>ij<12,2>",
		},
		{
			name:             "A match, almost too long",
			input:            []byte("XXXabcdXXX"),
			minMatchLen:      3,
			maxMatchLen:      3,
			maxSearchBuffLen: 255,
			wantValuesRepr:   "XXXabcd<7,3>",
		},
		{
			name:             "A match, too long but is not consumed",
			input:            []byte("XXXXabcdXXXX"),
			minMatchLen:      3,
			maxMatchLen:      3,
			maxSearchBuffLen: 255,
			wantValuesRepr:   "XXXXabcd<8,3>X",
		},
		{
			name:             "A match, outside search buffer",
			input:            []byte("XXXabcdefXXX"),
			minMatchLen:      3,
			maxMatchLen:      255,
			maxSearchBuffLen: 4,
			wantValuesRepr:   "XXXabcdefXXX",
		},
		{
			name:             "A match, almost outside search buffer",
			input:            []byte("XXXaXXX"),
			minMatchLen:      3,
			maxMatchLen:      255,
			maxSearchBuffLen: 4,
			wantValuesRepr:   "XXXa<4,3>",
		},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Run tests in parallel for efficiency

			values := BytesToValues(tt.input, tt.minMatchLen, tt.maxMatchLen, tt.maxSearchBuffLen)

			// Generate the string representation of values
			var valuesRepr string
			for _, v := range values {
				valuesRepr += fmt.Sprintf("%v", v)
			}

			if valuesRepr != tt.wantValuesRepr {
				t.Errorf("BytesToValues() = '%s'; want '%s'", valuesRepr, tt.wantValuesRepr)
			}
		})
	}
}

// Test_ValuesToBytes tests the ValuesToBytes function.
// It verifies that the function correctly converts a series of Value representations back into a byte slice,
// ensuring that the original data is accurately reconstructed.
func Test_ValuesToBytes(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "Only literals",
			input: []byte("abcdefghijkl"),
		},
		{
			name:  "Empty input",
			input: []byte(""),
		},
		{
			name:  "Single match",
			input: []byte("XXXaaaXXX"), // Expected to have a match like "XXXaaa<6,3>"
		},
		{
			name:  "Multiple matches",
			input: []byte("XXXabXXXcdXXXijXXX"),
		},
		{
			name:  "Repeated characters",
			input: []byte("XXXXXXXXXXXXXXXXXXXXXXX"),
		},
	}

	for _, tt := range tests {
		tt := tt // Capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Run tests in parallel for efficiency

			// Convert bytes to values with specified parameters
			values := BytesToValues(tt.input, 255, 255, 3)

			// Convert values back to bytes
			got := ValuesToBytes(values)

			if string(got) != string(tt.input) {
				t.Errorf("ValuesToBytes() = '%s'; want '%s'", string(got), string(tt.input))
			}
		})
	}
}

// Values is a global variable used in benchmarking to prevent compiler optimizations.
// It holds the result of BytesToValues during the benchmark.
var Values []Value

// Benchmark_ValuesToBytes benchmarks the performance of the BytesToValues function.
// It measures how efficiently the function can convert a large byte slice into Value representations.
func Benchmark_ValuesToBytes(b *testing.B) {
	// Generate 1000 random bytes for benchmarking
	randomBytes := make([]byte, 1000)
	_, err := rand.Read(randomBytes)
	if err != nil {
		b.Fatalf("Failed to generate random bytes: %v", err)
	}

	b.ResetTimer() // Reset the timer to exclude setup time

	for n := 0; n < b.N; n++ {
		Values = BytesToValues(randomBytes, 4, 255, 4096)
	}
}
