// Package ingests log files
package ingest

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	"gopkg.in/cheggaaa/pb.v1"

	"github.com/jasonrogena/gonx"
	"github.com/jasonrogena/log-analyse/sqlite"
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
	fmt.Printf("About to nom nom %d lines \n", noLines)

	if err != nil {
		log.Fatal(err)
		return err
	}

	db, err := sqlite.Connect()
	defer db.Close()

	if err != nil {
		log.Fatal(err)
		return err
	}

	startTime := time.Now().Unix()
	fileUUID, err := sqlite.Insert(
		db,
		"log_file",
		"uuid",
		[]string{"path", "no_lines", "start_time"},
		[]string{l.path, strconv.Itoa(noLines), strconv.FormatInt(startTime, 10)})
	if err != nil {
		log.Fatal(err)
		return err
	}
	l.uuid = fileUUID

	// Read the file line by line
	logFile.Seek(0, 0)
	lineNo := 0
	scanner := bufio.NewScanner(logFile)
	scanner.Split(bufio.ScanLines)
	nginxParser := gonx.NewParser(l.format)
	progress := pb.StartNew(noLines)
	for scanner.Scan() {
		lineNo = lineNo + 1
		logLine := scanner.Text()
		_, err = l.writeLine(db, nginxParser, &l, logLine, lineNo)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		}
		progress.Increment()
	}
	endTime := time.Now().Unix()
	timeDiff := float64(endTime-startTime) / 60.0
	progress.FinishPrint("Done ingesting " + l.path + " in " + strconv.FormatFloat(timeDiff, 'f', 2, 64) + "min")

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
