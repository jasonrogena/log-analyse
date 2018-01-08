// Package ingests log files
package ingest

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"
)

func ingestOneOff(accessLog Log) error {
	file, err := os.Open(accessLog.path)
	defer file.Close()

	if err != nil {
		log.Fatal(err)
		return err
	}

	// Get the number of lines
	noLines, err := getNumberLines(file)

	if err != nil {
		log.Fatal(err)
		return err
	}

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		logLine := scanner.Text()
		lineUUID, err := writeLine(accessLog, logLine)

		// Update progress interface
	}

	return nil
}

func getNumberLines(file *File) (lines int, err error) {
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
