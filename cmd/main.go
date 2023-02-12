package main

import (
	"fmt"
	"os"

	"github.com/akamensky/argparse"
)

func main() {
	// Parse command line arguments
	var outDir *string
	{
		// Create new argParser object
		argParser := argparse.NewParser("go-to-rust", "Outputs Rust files from Go files")
		// Create string flag
		outDir = argParser.String("o", "out-dir", &argparse.Options{Required: true, Help: "Output directory"})
		// Parse input
		err := argParser.Parse(os.Args)
		if err != nil {
			// In case of error print error and print usage
			// This can also be done by passing -h or --help flags
			fmt.Print(argParser.Usage(err))
		}
		// Finally print the collected string
		fmt.Println(*outDir)
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

	{
		data := list_nodes(src1)
		writeToFile("node_list.txt", data)
	}

	// {
	// 	data := node_tree(src1)
	// 	writeToFile("node_tree.txt", data)
	// }

	{
		data := parse_source(src1)
		writeToFile("output.txt", data)
	}
}
