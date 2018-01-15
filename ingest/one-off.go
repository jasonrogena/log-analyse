// Package ingests log files
package ingest

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/satyrius/gonx"
)

func ingestOneOff(l Log) error {
	logFile, err := os.Open(l.path)
	defer logFile.Close()

	if err != nil {
		log.Fatal(err)
		return err
	}

	// Get the number of lines
	noLines, _ := getNumberLines(logFile)
	fmt.Printf("Number of lines %d\n", noLines)

	if err != nil {
		log.Fatal(err)
		return err
	}

	// Read the file line by line
	logFile.Seek(0, 0)
	lineNo := 0
	scanner := bufio.NewScanner(logFile)
	scanner.Split(bufio.ScanLines)
	nginxParser := gonx.NewParser(l.format)
	for scanner.Scan() {
		fmt.Println("*********************************************************")
		lineNo = lineNo + 1
		fmt.Printf("Percent %f\n", (float64(lineNo)/float64(noLines))*100)
		logLine := scanner.Text()
		_, _ = l.writeLine(nginxParser, logLine, lineNo)

		// Update progress interface
		fmt.Println("#########################################################")
	}

	return nil
}

func getNumberLines(file *os.File) (lines int, err error) {
	buf := make([]byte, 1024)
	lines = 0
	for {
		readBytes, err := file.Read(buf)

		if err != nil {
			if readBytes == 0 && err == io.EOF {
				err = nil
			}
			return lines, err
		}
		lines += bytes.Count(buf[:readBytes], []byte{'\n'})
	}
}
