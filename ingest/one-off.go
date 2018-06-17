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
	"github.com/jasonrogena/log-analyse/config"
	"github.com/jasonrogena/log-analyse/sqlite"
	"github.com/jasonrogena/log-analyse/types"
)

func ingestOneOff(l types.Log, digesters []digester) error {
	logFile, err := os.Open(l.Path)
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
	defer sqlite.Disconnect(db)

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
		[]string{l.Path, strconv.Itoa(noLines), strconv.FormatInt(startTime, 10)},
		true)
	if err != nil {
		log.Fatal(err)
		return err
	}
	l.UUID = fileUUID

	// Initialize the goroutine pool
	noCPUs := runtime.NumCPU()
	if noLines < noCPUs {
		noCPUs = noLines
	}
	conf, cfgErr := config.GetConfig()
	if cfgErr != nil {
		log.Fatal(cfgErr)
		return cfgErr
	}

	// Create the goroutines
	fmt.Printf("Unleashing %d bunnies to nom nom the %d log lines\n", noCPUs, noLines)
	progressBar := pb.StartNew(noLines)
	linesWG := new(sync.WaitGroup)
	linesWG.Add(noCPUs)

	fieldsWG := new(sync.WaitGroup)
	fieldsWG.Add(1)
	logFieldsChan := make(chan *types.Field, 10000)
	go startDigestingFields(fieldsWG, &conf, digesters, logFieldsChan)
	logLines := make(chan types.Line, 100)
	for i := 0; i < noCPUs; i++ {
		go processLog(linesWG, db, &conf, &l, logLines, noLines, progressBar, logFieldsChan)
	}

	// Send log lines to logLines channel
	logFile.Seek(0, 0)
	var lineNo int64
	scanner := bufio.NewScanner(logFile)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		lineNo = lineNo + 1
		curLine := types.Line{Value: scanner.Text(), LineNo: lineNo}
		logLines <- curLine
	}
	close(logLines)

	linesWG.Wait() // wait for all the lines to be processed
	close(logFieldsChan)
	fieldsWG.Wait() // wait for all fields to be absorbed

	endTime := time.Now().Unix()
	timeDiff := float64(endTime-startTime) / 60.0
	progressBar.FinishPrint("Done ingesting " + l.Path + " in " + strconv.FormatFloat(timeDiff, 'f', 2, 64) + "min")

	if conf.Ingest.PiggyBackDigest {
		// Start process of absorption in all digesters
		noCPUs := runtime.NumCPU()
		if len(digesters) < noCPUs {
			noCPUs = len(digesters)
		}

		progressBar = pb.StartNew(len(digesters))
		startTime = time.Now().Unix()

		digestWG := new(sync.WaitGroup)
		digestWG.Add(noCPUs)
		digesterChan := make(chan digester, 10)

		for i := 0; i < noCPUs; i++ {
			go startDigesterAbsorb(digestWG, digesterChan, &conf, progressBar)
		}

		// Send the digesters to the waiting threads
		for _, curDigester := range digesters {
			digesterChan <- curDigester
		}
		close(digesterChan)

		digestWG.Wait()
		endTime = time.Now().Unix()
		timeDiff = float64(endTime-startTime) / 60.0
		progressBar.FinishPrint("Done digesting in " + strconv.FormatFloat(timeDiff, 'f', 2, 64) + "min")
	}

	return nil
}

func processLog(wg *sync.WaitGroup, db *sql.DB, cfg *config.Config, logFile *types.Log, logLines <-chan types.Line, noLines int, progressBar *pb.ProgressBar, logFields chan *types.Field) {
	defer wg.Done()

	nginxParser := gonx.NewParser(logFile.Format)

	// Read lines out of the logLines channel
	for line := range logLines {
		_, ingLineErr := writeLine(db, cfg, nginxParser, logFile, line, logFields)
		if ingLineErr != nil {
			fmt.Fprintln(os.Stderr, ingLineErr.Error())
		}
		progressBar.Increment()
	}
}

func startDigesterAbsorb(wg *sync.WaitGroup, digesters <-chan digester, cfg *config.Config, progressBar *pb.ProgressBar) {
	defer wg.Done()

	for curDigester := range digesters {
		digestErr := curDigester.Absorb(cfg)
		if digestErr != nil {
			fmt.Fprintln(os.Stderr, digestErr.Error())
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
