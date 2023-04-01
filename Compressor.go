package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
)

func main() {
	// Check if a filename was provided as an argument
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go <filename>")
		return
	}

	// Open the input file
	inFilename := os.Args[1]
	inFile, err := os.Open(inFilename)
	if err != nil {
		fmt.Println("Error opening input file:", err)
		return
	}
	defer inFile.Close()

	// Create the output file
	outFilename := inFilename + ".gz"
	outFile, err := os.Create(outFilename)
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer outFile.Close()

	// Create a gzip writer
	gzipWriter := gzip.NewWriter(outFile)
	defer gzipWriter.Close()

	// Copy the input file to the gzip writer
	_, err = io.Copy(gzipWriter, inFile)
	if err != nil {
		fmt.Println("Error compressing file:", err)
		return
	}

	fmt.Println("File compressed successfully:", outFilename)
}
