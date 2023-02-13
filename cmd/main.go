package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/akamensky/argparse"
)

func main() {
	// Parse command line arguments
	var inDir, outDir, fileName *string
	{
		// Create new argParser object
		argParser := argparse.NewParser("go-to-rust", "Outputs Rust files from Go files")
		// Create flags
		inDir = argParser.String("i", "in-dir", &argparse.Options{Required: true, Help: "Input directory"})
		outDir = argParser.String("o", "out-dir", &argparse.Options{Required: true, Help: "Output directory"})
		fileName = argParser.String("f", "file-name", &argparse.Options{Required: true, Help: "File name"})
		// Parse input
		err := argParser.Parse(os.Args)
		if err != nil {
			// In case of error print error and print usage
			// This can also be done by passing -h or --help flags
			fmt.Print(argParser.Usage(err))
			os.Exit(1)
		}
	}

	// Read input file
	readInput := func(fileName string) string {
		srcBytes, err := os.ReadFile(fmt.Sprintf("%v/%v", *inDir, fileName))
		if err != nil {
			panic(err)
		}
		return string(srcBytes)
	}

	// Write output to file
	writeToFile := func(fileName string, data string) {
		// Open the file for writing
		file, err := os.Create(fmt.Sprintf("%v/%v", *outDir, fileName))
		if err != nil {
			fmt.Println("Failed to create file:", err)
			return
		}
		defer file.Close()

		// Write some data to the file
		_, err = file.WriteString(data)
		if err != nil {
			fmt.Println("Failed to write to file:", err)
			return
		}

		fmt.Println("Data written to file:", fileName)
	}

	src := readInput(*fileName)
	filePrefix := strings.Split(*fileName, ".")[0]

	// Write _node_list.txt file to aid in troubleshooting
	{
		data := list_nodes(src)
		writeToFile(filePrefix+"_node_list.txt", data)
	}

	// Write AST file to aid in troubleshooting
	// {
	// 	data := node_tree(src)
	// 	writeToFile(filePrefix+"_node_tree.txt", data)
	// }

	// Write .rx file
	{
		data := parseToRust(src)
		writeToFile(filePrefix+".rs", data)
	}
}
