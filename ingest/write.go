package ingest

import (
	"fmt"
	"os"

	"github.com/jasonrogena/gonx"
)

func (log Log) writeLine(parser *gonx.Parser, line string, lineNo int) (uuid string, err error) {
	entry, err := parser.ParseString(line)
	if err != nil {
		return
	}

	// TODO: insert log line to database

	// Save the entry
	for _, curField := range fields {
		switch curField.typ {
		case typ_string:
			strVal, err := entry.Field(curField.name)
			if err == nil {
				fmt.Fprintf(os.Stdout, "Value for %q is %q\n", curField.name, strVal)
				// TODO: insert log field to database
			}
		case typ_float:
			fltVal, err := entry.FloatField(curField.name)
			if err == nil {
				fmt.Fprintf(os.Stdout, "Value for %q is %f\n", curField.name, fltVal)
				// TODO: insert log field to database
			}
		}
	}
	return
}
