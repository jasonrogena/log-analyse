package ingest

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jasonrogena/gonx"
	"github.com/jasonrogena/log-analyse/sqlite"
)

func (log Log) writeLine(db *sql.DB, parser *gonx.Parser, logFile *Log, line string, lineNo int) (string, error) {
	entry, parseErr := parser.ParseString(line)
	if parseErr != nil {
		return "", parseErr
	}

	startTime := time.Now().Unix()
	lineUUID, lineInstErr := sqlite.Insert(
		db,
		"log_line",
		"uuid",
		[]string{"line_no", "value", "start_time", "log_file_uuid"},
		[]string{strconv.Itoa(lineNo), line, strconv.FormatInt(startTime, 10), logFile.uuid})
	fmt.Printf("Log line UUID %q\n", lineUUID)
	if lineInstErr != nil {
		return "", lineInstErr
	}

	// Save the entry
	for _, curField := range fields {
		switch curField.typ {
		case typ_string:
			strVal, strErr := entry.Field(curField.name)
			if strErr == nil {
				st := time.Now().Unix()
				_, fieldInstErr := sqlite.Insert(
					db,
					"log_field",
					"uuid",
					[]string{"field_type", "value_type", "value", "start_time", "log_line_uuid"},
					[]string{curField.name, "string", strVal, strconv.FormatInt(st, 10), lineUUID})
				if fieldInstErr != nil {
					fmt.Fprintln(os.Stderr, fieldInstErr.Error())
				}
			}
		case typ_float:
			fltVal, fltErr := entry.FloatField(curField.name)
			if fltErr == nil {
				st := time.Now().Unix()
				_, fieldInstErr := sqlite.Insert(
					db,
					"log_field",
					"uuid",
					[]string{"field_type", "value_type", "value", "start_time", "log_line_uuid"},
					[]string{curField.name, "float", strconv.FormatFloat(fltVal, 'E', -1, 64), strconv.FormatInt(st, 10), lineUUID})
				if fieldInstErr != nil {
					fmt.Fprintln(os.Stderr, fieldInstErr.Error())
				}
			}
		}
	}
	return lineUUID, nil
}