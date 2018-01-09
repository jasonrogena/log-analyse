package main

import "github.com/jasonrogena/log-analyse/ingest"
import "fmt"

func process(command string, args []string) {
	switch command {
	case ingest.Name:
		ingest.Process(args)
	default:
		printHelp()
	}
}

func printHelp() {
	fmt.Println("Unknown command")
}
