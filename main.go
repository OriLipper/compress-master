// main.go
// Package main implements a compression and decompression tool using LZ77 and Huffman coding algorithms.
// It provides functionalities to compress files by replacing repeated sequences with pointers and
// encoding the result using Huffman coding for efficient storage. Additionally, it can decompress
// the encoded files back to their original form.
//
// The program supports various command-line options for configuring compression parameters,
// generating diagnostic outputs like Huffman tree visualizations, and profiling performance.
package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime/pprof"
	"strings"
	"time"
)

func compress(
	source io.Reader,
	sink io.Writer,
	minMatch byte,
	maxMatch byte,
	searchSize uint16,

	graphf io.Writer,
	lzf io.Writer,
) {
	log.Printf("Config: min-match=%d, max-match=%d, search-size=%d\n", minMatch, maxMatch, searchSize)
	input, err := ioutil.ReadAll(source)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Input size (bytes): %d\n", len(input))
	// LZ coding.
	values := BytesToValues(input, minMatch, maxMatch, searchSize)
	// Optionally write LZ77 representation
	if lzf != ioutil.Discard {
		for _, v := range values {
			fmt.Fprintf(lzf, "%v", v)
		}
	}
	// Huffman coding.
	root := constructHuffmanTree(values)
	root.DumpGraphviz(graphf)
	codeTable := createCodeTable(root, Code{})
	// Write binary representation.
	bw := NewBinaryWriter(sink, codeTable)
	bw.Write(values)
}

func decompress(source io.Reader, sink io.Writer) {
	br := NewBinaryReader(source)
	newVals := br.Read()
	_, err := sink.Write(ValuesToBytes(newVals))
	if err != nil {
		log.Fatal(err)
	}
}

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] <filename>\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	var (
		minMatch       uint
		maxMatch       uint
		searchSize     uint
		verbose        bool
		graphvizPath   string
		lzPath         string
		cpuProfilePath string
	)

	// Define command-line flags.
	flag.BoolVar(&verbose, "verbose", false, "Display log messages")
	compressMode := flag.Bool("compress", true, "Run the program in compression mode")
	flag.StringVar(&graphvizPath, "graphviz", "", "Write Graphviz Huffman tree representation to file")
	flag.StringVar(&lzPath, "lz", "", "Write LZ77 representation to file")
	flag.StringVar(&cpuProfilePath, "cpuprofile", "", "Write CPU profile to file")
	flag.String("name", "", "Name for the output file (compressed or decompressed)")
	flag.UintVar(&minMatch, "min-match", 4, "Minimum match size for LZ77 algorithm")
	flag.UintVar(&maxMatch, "max-match", 255, "Maximum match size for LZ77 algorithm (upper limit is 255)")
	flag.UintVar(&searchSize, "search-size", 4096, "Size of the search window for LZ77 algorithm (upper limit is 65535)")

	// Customize the usage message.
	flag.Usage = Usage

	// Parse the flags.
	flag.Parse()

	// Validate the number of positional arguments.
	if flag.NArg() != 1 {
		Usage()
	}

	// Retrieve the filename from positional arguments.
	filePath := flag.Arg(0)

	// Configure logging based on the verbose flag.
	if !verbose {
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetOutput(os.Stdout)
		log.Printf("Running %s in verbose mode\n", os.Args[0])
	}

	// Start CPU profiling if the cpuprofile flag is set.
	if cpuProfilePath != "" {
		log.Printf("Starting CPU profiling: %s\n", cpuProfilePath)
		f, err := os.Create(cpuProfilePath)
		if err != nil {
			log.Fatalf("Failed to create CPU profile file: %v", err)
		}
		defer func() {
			pprof.StopCPUProfile()
			f.Close()
			log.Println("CPU profiling stopped.")
		}()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatalf("Failed to start CPU profile: %v", err)
		}
	}

	// Open the Graphviz writer if the graphviz flag is set.
	var graphf io.Writer
	if graphvizPath != "" {
		log.Printf("Creating Graphviz Huffman tree representation: %s\n", graphvizPath)
		graphfFile, err := os.Create(graphvizPath)
		if err != nil {
			log.Fatalf("Failed to create Graphviz file: %v", err)
		}
		defer graphfFile.Close()
		graphf = graphfFile
	} else {
		graphf = ioutil.Discard
	}

	// Open the LZ77 writer if the lz flag is set.
	var lzf io.Writer
	if lzPath != "" {
		log.Printf("Creating LZ77 representation: %s\n", lzPath)
		lzfFile, err := os.Create(lzPath)
		if err != nil {
			log.Fatalf("Failed to create LZ77 file: %v", err)
		}
		defer lzfFile.Close()
		lzf = lzfFile
	} else {
		lzf = ioutil.Discard
	}

	// Open the input file.
	inputFile, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open input file '%s': %v", filePath, err)
	}
	defer inputFile.Close()

	if *compressMode {
		// Compression mode.
		log.Printf("Compressing file: %s\n", filePath)

		// Determine the output filename.
		outputName := flag.Lookup("name").Value.String()
		if outputName == "" {
			outputName = filePath + ".compressed"
		}

		// Open the output file for writing compressed data.
		outputFile, err := os.Create(outputName)
		if err != nil {
			log.Fatalf("Failed to create output file '%s': %v", outputName, err)
		}
		defer outputFile.Close()

		// Get the original file size for compression ratio calculation.
		originalFileSize := getFileSize(filePath)

		// Start the compression process and measure the time taken.
		startTime := time.Now()
		compress(inputFile, outputFile, byte(minMatch), byte(maxMatch), uint16(searchSize), graphf, lzf)
		elapsedTime := time.Since(startTime)

		// Get the compressed file size.
		compressedFileSize := getFileSize(outputName)

		// Log compression statistics.
		log.Printf("Compression Time Elapsed: %s\n", elapsedTime)
		log.Printf("Original File Size: %d bytes\n", originalFileSize)
		log.Printf("Compressed File Size: %d bytes\n", compressedFileSize)
		if compressedFileSize > 0 {
			log.Printf("Compression Ratio: %.2f\n", float64(originalFileSize)/float64(compressedFileSize))
		} else {
			log.Println("Compressed file size is zero; cannot compute compression ratio.")
		}
	} else {
		// Decompression mode.
		log.Printf("Decompressing file: %s\n", filePath)

		// Determine the output filename.
		outputName := flag.Lookup("name").Value.String()
		if outputName == "" {
			// Attempt to remove the ".compressed" suffix if present.
			if strings.HasSuffix(filePath, ".compressed") {
				outputName = filePath[:len(filePath)-11] + ".decompressed"
			} else {
				outputName = filePath + ".decompressed"
			}
		}

		// Open the output file for writing decompressed data.
		outputFile, err := os.Create(outputName)
		if err != nil {
			log.Fatalf("Failed to create output file '%s': %v", outputName, err)
		}
		defer outputFile.Close()

		// Start the decompression process and measure the time taken.
		startTime := time.Now()
		decompress(inputFile, outputFile)
		elapsedTime := time.Since(startTime)

		// Log decompression statistics.
		log.Printf("Decompression Time Elapsed: %s\n", elapsedTime)
		log.Printf("Decompressed File Size: %d bytes\n", getFileSize(outputName))
	}
}
