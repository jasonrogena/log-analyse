package ingest

import (
	"fmt"
)

const Name = "ingest"
const oneOff = "one-off"

type Log struct {
	path string
}

func Process(args []string) {
	if len(args) == 2 {
		ingestType := args[0]
		log := Log{path: args[1]}
		switch ingestType {
		case oneOff:
			ingestOneOff(log)
		}
	} else {
		printHelp()
	}
}

func printHelp() {
	fmt.Println("Invalid ingest arguements")
}
