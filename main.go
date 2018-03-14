package main

import (
	"fmt"
	"os"

	"github.com/jasonrogena/log-analyse/ingest"
)

func main() {
	if len(os.Args) > 1 {
		command := os.Args[1]
		args := os.Args[2:]
		process(command, args)
	} else {
		fmt.Printf("Command not specified\n\n")
		printHelp()
	}
}

func process(command string, args []string) {
	switch command {
	case ingest.Name:
		ingest.Process(args)
	default:
		fmt.Printf("Unknown command\n\n")
		printHelp()
	}
}

func printHelp() {
	fmt.Printf("Usage %s <command>\n\n", os.Args[0])
	fmt.Printf("Common commands:\n")
	fmt.Printf("\tingest\t\tGets contents from a NGINX access log\n\n")
}
