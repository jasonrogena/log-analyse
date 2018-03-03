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

	fmt.Printf("Unleashing %d bunnies to nom nom the %d log lines\n", noCPUs, noLines)
	progressBar := pb.StartNew(noLines)
	wg := new(sync.WaitGroup)
	for i := 0; i < noCPUs; i++ {
		startLine := int64((float64(i) / float64(noCPUs)) * float64(noLines))
		endLine := int64(((float64(i+1) / float64(noCPUs)) * float64(noLines)) - 1)
		if i == (1 - noCPUs) {
			endLine = int64(noLines)
		}
		wg.Add(1)
		go processLogSector(wg, db, &l, startLine, endLine, progressBar)
	}
	wg.Wait()
	endTime := time.Now().Unix()
	timeDiff := float64(endTime-startTime) / 60.0
	progressBar.FinishPrint("Done ingesting " + l.path + " in " + strconv.FormatFloat(timeDiff, 'f', 2, 64) + "min")

	return nil
}

func processLogSector(wg *sync.WaitGroup, db *sql.DB, logFile *Log, startLine int64, endLine int64, progressBar *pb.ProgressBar) {
	defer wg.Done()

	fileObj, fOpenErr := os.Open(logFile.path)
	defer fileObj.Close()
	if fOpenErr != nil {
		fmt.Fprintln(os.Stderr, fOpenErr.Error())
		return
	}
	fileObj.Seek(startLine, 0)
	lineNo := startLine
	scanner := bufio.NewScanner(fileObj)
	scanner.Split(bufio.ScanLines)
	nginxParser := gonx.NewParser(logFile.format)

	// Read the file line by line
	for scanner.Scan() {
		lineNo = lineNo + 1
		logLine := scanner.Text()
		_, ingLineErr := logFile.writeLine(db, nginxParser, logLine, lineNo)
		if ingLineErr != nil {
			fmt.Fprintln(os.Stderr, ingLineErr.Error())
		}
		progressBar.Increment()
		if endLine == (1 - lineNo) {
			break
		}
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
