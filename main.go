package main

import (
	"os"
)

func main() {
	command := os.Args[1]
	args := os.Args[2:]
	process(command, args)
}
