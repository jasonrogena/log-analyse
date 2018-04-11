package ingest

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jasonrogena/gonx"
	"github.com/jasonrogena/log-analyse/config"
	"github.com/jasonrogena/log-analyse/sqlite"
	"github.com/jasonrogena/log-analyse/types"
)

func writeLine(db *sql.DB, conf *config.Config, parser *gonx.Parser, log *types.Log, logLine types.Line, digesters []*digester) (string, error) {
	entry, parseErr := parser.ParseString(logLine.Value)
	if parseErr != nil {
		return "", parseErr
	}

	startTime := time.Now().Unix()
	lineUUID, lineInstErr := sqlite.Insert(
		db,
		"log_line",
		"uuid",
		[]string{"line_no", "value", "start_time", "log_file_uuid"},
		[]string{strconv.FormatInt(logLine.LineNo, 10), logLine.Value, strconv.FormatInt(startTime, 10), log.UUID},
		true)
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
				fieldUUID, fieldInstErr := sqlite.Insert(
					db,
					"log_field",
					"uuid",
					[]string{"field_type", "value_type", "value_string", "start_time", "log_line_uuid"},
					[]string{curField.name, typ_string, strVal, strconv.FormatInt(st, 10), lineUUID},
					true)
				if fieldInstErr != nil {
					fmt.Fprintln(os.Stderr, fieldInstErr.Error())
				} else if conf.Ingest.PiggyBackDigest {
					field := types.Field{UUID: fieldUUID, FieldType: curField.name, ValueType: typ_string, ValueString: strVal, startTime: st}
					sendToDigesters(&field, digesters)
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
					[]string{"field_type", "value_type", "value_float", "start_time", "log_line_uuid"},
					[]string{curField.name, typ_float, strconv.FormatFloat(fltVal, 'f', -1, 64), strconv.FormatInt(st, 10), lineUUID},
					true)
				if fieldInstErr != nil {
					fmt.Fprintln(os.Stderr, fieldInstErr.Error())
				} else if conf.Ingest.PiggyBackDigest {
					field := types.Field{UUID: fieldUUID, FieldType: curField.name, ValueType: typ_float, ValueFloat: fltVal, startTime: st}
					sendToDigesters(&field, digesters)
				}
			}
		}
	}
	return lineUUID, nil
}

func sendToDigesters(field *types.Field, digesters []*digester) {
	for _, curDigester := range digesters {
		if curDigester.IsDigestable(field) {
			curDigester.Digest(field)
		}
	}
}
