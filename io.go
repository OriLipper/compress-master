// io.go
// Package main provides functionality for writing and reading compressed data using a binary format.
// It defines BinaryWriter and BinaryReader types that handle the serialization and deserialization
// of Value slices based on a provided CodeTable. The package leverages bit-level IO operations
// to efficiently encode literals and pointers as part of an LZ77-like compression algorithm.

package main

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/icza/bitio"
)

// BinaryWriter is responsible for serializing Value slices into a binary format.
// It utilizes a CodeTable to encode literals and pointers efficiently.
type BinaryWriter struct {
	w         *bitio.Writer // Bit-level writer for output operations.
	codeTable CodeTable     // Mapping of bytes to their corresponding binary codes.
}

// NewBinaryWriter creates and returns a new BinaryWriter.
// Parameters:
// - writer: An io.Writer where the binary data will be written.
// - codeTable: A CodeTable that defines the encoding scheme for literals and pointers.
func NewBinaryWriter(writer io.Writer, codeTable CodeTable) BinaryWriter {
	bitWriter := bitio.NewWriter(writer)
	return BinaryWriter{
		w:         bitWriter,
		codeTable: codeTable,
	}
}

// Write serializes a slice of Value instances into binary format.
// It writes the code table first, followed by each Value's data.
// Parameters:
// - values: A slice of Value instances to be serialized.
func (bw *BinaryWriter) Write(values []Value) {
	// Write the code table to the binary stream.
	bw.writeTable()

	// Iterate over each Value and serialize it.
	for _, v := range values {
		// Write the IsLiteral flag as a single bit.
		if err := bw.w.WriteBool(v.IsLiteral); err != nil {
			panic("BinaryWriter.Write: failed to write IsLiteral flag")
		}

		if v.IsLiteral {
			// For literals, retrieve the corresponding code and bit length.
			code, bitLen := bw.getCodeForValue(v.GetLiteralBinary())
			// Write the literal's code as bits.
			if err := bw.w.WriteBits(code, bitLen); err != nil {
				panic("BinaryWriter.Write: failed to write literal bits")
			}
		} else {
			// For pointers, serialize each byte of the pointer.
			pointerBytes := v.GetPointerBinary()
			for _, b := range pointerBytes {
				code, bitLen := bw.getCodeForValue(b)
				// Write each byte of the pointer as bits.
				if err := bw.w.WriteBits(code, bitLen); err != nil {
					panic("BinaryWriter.Write: failed to write pointer bits")
				}
			}
		}
	}

	// Close the bit writer to flush any remaining bits.
	if err := bw.w.Close(); err != nil {
		panic("BinaryWriter.Write: failed to close bit writer")
	}
}

// writeTable serializes the CodeTable into the binary stream.
// It writes the number of table entries followed by each (value, bit length, code) triplet.
func (bw *BinaryWriter) writeTable() {
	// Ensure the CodeTable is not empty.
	if len(bw.codeTable) == 0 {
		panic("BinaryWriter.writeTable: code table has zero length")
	}

	// Write the number of elements in the CodeTable as 8 bits.
	// Subtract 1 to prevent overflow when the table size is 256.
	if err := bw.w.WriteBits(uint64(len(bw.codeTable)-1), 8); err != nil {
		panic("BinaryWriter.writeTable: failed to write table size")
	}

	// Iterate over the CodeTable and write each entry.
	for byteVal, code := range bw.codeTable {
		// Write the byte value (8 bits).
		if err := bw.w.WriteBits(uint64(byteVal), 8); err != nil {
			panic("BinaryWriter.writeTable: failed to write byte value")
		}

		// Write the number of bits for the code (8 bits).
		if err := bw.w.WriteBits(uint64(code.bits), 8); err != nil {
			panic("BinaryWriter.writeTable: failed to write code bit length")
		}

		// Write the actual code (variable bits as defined by code.bits).
		if err := bw.w.WriteBits(uint64(code.c), code.bits); err != nil {
			panic("BinaryWriter.writeTable: failed to write code bits")
		}
	}
}

// getCodeForValue retrieves the binary code and its bit length for a given byte value.
// Parameters:
// - val: The byte value to retrieve the code for.
// Returns:
// - code: The binary code as a uint64.
// - bitLen: The number of bits in the code.
func (bw *BinaryWriter) getCodeForValue(val byte) (uint64, byte) {
	codeEntry, exists := bw.codeTable[val]
	if !exists {
		panic("BinaryWriter.getCodeForValue: code not found for value")
	}
	return uint64(codeEntry.c), codeEntry.bits
}

// BinaryReader is responsible for deserializing binary data into Value slices.
// It reads the code table first, then reconstructs each Value based on the serialized data.
type BinaryReader struct {
	r        *bitio.Reader // Bit-level reader for input operations.
	valTable map[Code]byte // Reverse mapping from codes to byte values.
}

// NewBinaryReader creates and returns a new BinaryReader.
// Parameters:
// - reader: An io.Reader from which the binary data will be read.
func NewBinaryReader(reader io.Reader) BinaryReader {
	bitReader := bitio.NewReader(reader)
	return BinaryReader{
		r: bitReader,
	}
}

// Read deserializes binary data into a slice of Value instances.
// It first reads the code table, then iterates through the binary stream to reconstruct each Value.
// Returns:
// - A slice of Value instances representing the decompressed data.
func (br *BinaryReader) Read() []Value {
	// Deserialize the code table.
	br.valTable = br.readTable()

	// Initialize a slice to hold the reconstructed Values.
	values := make([]Value, 0)

	// Continuously consume Values until EOF is reached.
	for {
		val, err := br.consumeValue()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break // End of binary stream reached.
			}
			panic("BinaryReader.Read: failed to consume value")
		}
		values = append(values, val)
	}

	return values
}

// readTable deserializes the CodeTable from the binary stream.
// It reads the number of table entries and then reads each (code, byte value) pair.
// Returns:
// - A map mapping Code structs to their corresponding byte values.
func (br *BinaryReader) readTable() map[Code]byte {
	valTable := make(map[Code]byte)

	// Read the number of elements in the table (8 bits).
	sizeBits, err := br.r.ReadBits(8)
	if err != nil {
		panic("BinaryReader.readTable: failed to read table size")
	}
	// Add 1 to account for the earlier subtraction during writing.
	size := sizeBits + 1

	// Iterate to read each table entry.
	for i := uint64(0); i < size; i++ {
		// Read the byte value (8 bits).
		valBits, err := br.r.ReadBits(8)
		if err != nil {
			panic("BinaryReader.readTable: failed to read byte value")
		}
		val := byte(valBits)

		// Read the number of bits in the code (8 bits).
		codeBits, err := br.r.ReadBits(8)
		if err != nil {
			panic("BinaryReader.readTable: failed to read code bit length")
		}
		codeLength := byte(codeBits)

		// Read the actual code based on the bit length.
		codeValue, err := br.r.ReadBits(codeLength)
		if err != nil {
			panic("BinaryReader.readTable: failed to read code bits")
		}
		code := Code{
			c:    codeValue,
			bits: codeLength,
		}

		// Populate the reverse mapping table.
		valTable[code] = val
	}

	return valTable
}

// consumeValue deserializes a single Value from the binary stream.
// It reads the IsLiteral flag and reconstructs either a literal or a pointer based on the flag.
// Returns:
// - A Value instance.
// - An error if the deserialization fails.
func (br *BinaryReader) consumeValue() (Value, error) {
	// Read the IsLiteral flag (1 bit).
	isLiteral, err := br.r.ReadBool()
	if err != nil {
		return Value{}, err
	}

	if isLiteral {
		// Deserialize a literal Value.
		literal, err := br.readMatch()
		if err != nil {
			return Value{}, err
		}
		return NewValue(true, literal, 0, 0), nil
	}

	// Deserialize a pointer Value.
	pointerBytes, err := br.readPointerMatches()
	if err != nil {
		return Value{}, err
	}
	return pointerMatchesToPointer(pointerBytes), nil
}

// readMatch deserializes a single byte value based on the code table.
// It reads bits until a matching code is found in the valTable.
// Returns:
// - The corresponding byte value.
// - An error if deserialization fails.
func (br *BinaryReader) readMatch() (byte, error) {
	currentCode := Code{}

	for {
		// Read the next bit and append it to the current code.
		bit, err := br.r.ReadBool()
		if err != nil {
			return 0, err
		}
		currentCode = addBit(currentCode, bit)

		// Check if the current code exists in the valTable.
		if val, exists := br.valTable[currentCode]; exists {
			return val, nil
		}
	}
}

// readPointerMatches deserializes the three bytes that make up a pointer Value.
// Returns:
// - A slice of three bytes representing the pointer.
// - An error if deserialization fails.
func (br *BinaryReader) readPointerMatches() ([]byte, error) {
	pointerBytes := make([]byte, 3)
	var err error

	// Deserialize each byte of the pointer.
	for i := 0; i < 3; i++ {
		pointerBytes[i], err = br.readMatch()
		if err != nil {
			return nil, err
		}
	}

	return pointerBytes, nil
}

// pointerMatchesToPointer converts a slice of three bytes into a pointer Value.
// The first two bytes represent the distance, and the third byte represents the length.
// Parameters:
// - bytes: A slice of three bytes representing the pointer.
// Returns:
// - A Value instance representing the pointer.
func pointerMatchesToPointer(bytes []byte) Value {
	// Decode the first two bytes as a big-endian uint16 for distance.
	distance := binary.BigEndian.Uint16(bytes[:2])
	// The third byte is the length of the match.
	length := bytes[2]

	return NewValue(false, 0, length, distance)
}
