// values.go
// Package main provides functionality for encoding and decoding data using a simplified LZ77 compression algorithm.
// It defines the Value type, which represents either a literal byte or a pointer to a previous sequence in the data.
// The package includes functions to convert between byte slices and Value slices, as well as utility functions to support these operations.

package main

import (
	"encoding/binary"
	"fmt"
	"log"
)

// Value represents an element in the LZ77 compression sequence.
// It can be either a literal byte or a pointer to a sequence of bytes previously seen.
type Value struct {
	IsLiteral bool // Indicates whether the Value is a literal byte.

	// Literal representation.
	val byte // The literal byte value.

	// Pointer representation.
	distance uint16 // The distance back from the current position to the start of the matching sequence.
	length   byte   // The length of the matching sequence.
}

// NewValue constructs a new Value instance.
// Parameters:
// - isLiteral: true if the Value is a literal byte, false if it is a pointer.
// - value: the literal byte value (ignored if isLiteral is false).
// - length: the length of the match (relevant if isLiteral is false).
// - distance: the distance back to the match (relevant if isLiteral is false).
func NewValue(isLiteral bool, value, length byte, distance uint16) Value {
	return Value{
		IsLiteral: isLiteral,
		val:       value,
		distance:  distance,
		length:    length,
	}
}

// String returns the string representation of the Value.
// For literals, it returns the character itself.
// For pointers, it returns a string in the format "<distance,length>".
func (v Value) String() string {
	if v.IsLiteral {
		return fmt.Sprintf("%c", v.val)
	}
	return fmt.Sprintf("<%d,%d>", v.distance, v.length)
}

// GetLiteralBinary returns the binary representation of a literal Value.
// It simply returns the literal byte.
func (v *Value) GetLiteralBinary() byte {
	return v.val
}

// GetPointerBinary returns the binary representation of a pointer Value.
// It serializes the distance and length into a byte slice using big-endian encoding.
// The first two bytes represent the distance, and the third byte represents the length.
func (v *Value) GetPointerBinary() []byte {
	bytes := make([]byte, 3)
	// Encode the distance as the first two bytes in big-endian order.
	binary.BigEndian.PutUint16(bytes, v.distance)
	// The third byte is the length.
	bytes[2] = v.length
	return bytes
}

// BytesToValues converts a byte slice into a slice of Value instances using LZ77 compression.
// It replaces sequences of bytes with pointers to previous occurrences where possible.
// Parameters:
// - input: the input byte slice to be compressed.
// - minMatchLen: the minimum length of a match to be considered for compression.
// - maxMatchLen: the maximum length of a match.
// - maxSearchBuffLen: the maximum length of the search buffer.
func BytesToValues(input []byte, minMatchLen, maxMatchLen byte, maxSearchBuffLen uint16) []Value {
	var (
		searchBuffStart  int
		lookaheadBuffEnd int
		matchPos         int
		matchLen         byte
		distance         uint16
	)

	// Preallocate the values slice with the length of input.
	// It is likely to be over-allocated, but slicing will adjust the final size.
	values := make([]Value, len(input))
	valueCounter := 0   // Tracks the number of values added.
	pointerCounter := 0 // Tracks the number of pointers used.

	for split := 0; split < len(input); split++ {
		// Define the boundaries of the search buffer.
		searchBuffStart = max(0, split-int(maxSearchBuffLen))
		// Define the end of the lookahead buffer.
		lookaheadBuffEnd = min(len(input), split+int(maxMatchLen))

		// Find the longest match position and length within the current buffers.
		matchPos, matchLen = getLongestMatchPosAndLen(
			input[searchBuffStart:split],
			input[split:lookaheadBuffEnd],
			minMatchLen,
		)

		if split > int(minMatchLen) && matchLen > 0 {
			// Calculate the distance from the current position to the match position.
			distance = uint16(split - (matchPos + searchBuffStart))
			// Create a pointer Value.
			values[valueCounter] = NewValue(false, 0, matchLen, distance)
			valueCounter++
			// Advance the split position by the length of the match minus one.
			split += int(matchLen) - 1
			pointerCounter++
		} else {
			// Create a literal Value.
			values[valueCounter] = NewValue(true, input[split], 1, 0)
			valueCounter++
		}
	}

	// Log the ratio of pointers to total values for diagnostic purposes.
	log.Printf("Pointers ratio: %.2f\n", float64(pointerCounter)/float64(valueCounter))
	// Return the slice of values up to the number of values added.
	return values[:valueCounter]
}

// getLongestMatchPosAndLen finds the position and length of the longest match between the text and the pattern.
// Parameters:
// - text: the search buffer where matches are sought.
// - pattern: the lookahead buffer where matches are compared.
// - minMatchLen: the minimum length of a match to be considered.
// Returns:
// - position: the starting index of the longest match in the text.
// - length: the length of the longest match.
func getLongestMatchPosAndLen(text, pattern []byte, minMatchLen byte) (int, byte) {
	if len(pattern) < int(minMatchLen) {
		return 0, 0
	}

	var (
		matchLen byte
		maxSoFar byte
		position int
	)

	// Find all starting indices in text where the first minMatchLen bytes of pattern match.
	minMatchStarts := getMatchIndex(text, pattern[:minMatchLen])

	for _, matchStart := range minMatchStarts {
		// Determine the length of the match starting at matchStart.
		matchLen = getMatchLen(text[matchStart:], pattern)
		// Update the longest match found so far.
		if matchLen >= minMatchLen && matchLen > maxSoFar {
			position = matchStart
			maxSoFar = matchLen
		}
	}

	return position, maxSoFar
}

// getMatchIndex returns a slice of starting indices in text where the pattern begins.
// Parameters:
// - text: the text to search within.
// - pattern: the byte pattern to search for.
// Returns:
// - A slice of integers representing the starting indices of matches.
func getMatchIndex(text, pattern []byte) []int {
	// If text is shorter than pattern or either is empty, there are no matches.
	if len(text) == 0 || len(text) < len(pattern) {
		return []int{}
	}

	// If pattern is empty, every index is a match.
	if len(pattern) == 0 {
		matchIndices := make([]int, len(text))
		for i := range text {
			matchIndices[i] = i
		}
		return matchIndices
	}

	matchIndices := []int{}
	// Iterate through text to find matches of pattern.
	for i := 0; i <= len(text)-len(pattern); i++ {
		// Optimized comparison: check first three bytes explicitly for speed.
		match := true
		for j := 0; j < len(pattern); j++ {
			if text[i+j] != pattern[j] {
				match = false
				break
			}
		}
		if match {
			matchIndices = append(matchIndices, i)
		}
	}
	return matchIndices
}

// getMatchLen returns the length of the longest match between two byte slices.
// It compares byte by byte up to the maximum possible match length or 255, whichever is smaller.
// Parameters:
// - a: the first byte slice.
// - b: the second byte slice.
// Returns:
// - The length of the longest matching sequence.
func getMatchLen(a, b []byte) byte {
	var matchLen byte
	// Determine the maximum possible match length.
	maxMatchLen := min(min(len(a), len(b)), 255)
	for i := 0; i < maxMatchLen; i++ {
		if a[i] == b[i] {
			matchLen++
		} else {
			break
		}
	}
	return matchLen
}

// ValuesToBytes converts a slice of Value instances back into a byte slice.
// It reconstructs the original data by replacing pointers with the corresponding byte sequences.
// Parameters:
// - values: the slice of Value instances to be converted.
// Returns:
// - A byte slice representing the reconstructed data.
func ValuesToBytes(values []Value) []byte {
	var from int
	bytesResult := make([]byte, 0, len(values)) // Preallocate with an estimated capacity.

	for _, v := range values {
		if v.IsLiteral {
			// Append the literal byte directly.
			bytesResult = append(bytesResult, v.val)
		} else {
			// Calculate the starting index from which to copy the bytes.
			from = len(bytesResult) - int(v.distance)
			// Append the matched sequence based on distance and length.
			bytesResult = append(bytesResult, bytesResult[from:from+int(v.length)]...)
		}
	}

	return bytesResult
}
