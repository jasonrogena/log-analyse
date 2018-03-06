// Package ingests log files
package ingest

import (
	"bufio"
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
	"sync"
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

	if err != nil {
		log.Fatal(err)
		return err
	}

	db, err := sqlite.Connect(true)
	defer db.Close()

	if err != nil {
		log.Fatal(err)
		return err
	}

	// Insert file into the database
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

	// Initialize the goroutine pool
	noCPUs := runtime.NumCPU()
	if noLines < noCPUs {
		noCPUs = noLines
	}

	// Create the goroutines
	fmt.Printf("Unleashing %d bunnies to nom nom the %d log lines\n", noCPUs, noLines)
	progressBar := pb.StartNew(noLines)
	wg := new(sync.WaitGroup)
	wg.Add(noCPUs)
	logLines := make(chan Line, 100)
	for i := 0; i < noCPUs; i++ {
		go processLog(wg, db, &l, logLines, progressBar)
	}

	// Send log lines to logLines channel
	logFile.Seek(0, 0)
	var lineNo int64
	scanner := bufio.NewScanner(logFile)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		lineNo = lineNo + 1
		curLine := Line{value: scanner.Text(), lineNo: lineNo}
		logLines <- curLine
	}
	close(logLines)

	wg.Wait()
	endTime := time.Now().Unix()
	timeDiff := float64(endTime-startTime) / 60.0
	progressBar.FinishPrint("Done ingesting " + l.path + " in " + strconv.FormatFloat(timeDiff, 'f', 2, 64) + "min")

	return nil
}

func processLog(wg *sync.WaitGroup, db *sql.DB, logFile *Log, logLines <-chan Line, progressBar *pb.ProgressBar) {
	defer wg.Done()

	nginxParser := gonx.NewParser(logFile.format)

	// Read lines out of the logLines channel
	for line := range logLines {
		_, ingLineErr := logFile.writeLine(db, nginxParser, line)
		if ingLineErr != nil {
			fmt.Fprintln(os.Stderr, ingLineErr.Error())
		}
		progressBar.Increment()
	}
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
