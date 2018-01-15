package ingest

import (
	"fmt"

	"github.com/satyrius/gonx"
)

func (log Log) writeLine(parser *gonx.Parser, line string, lineNo int) (uuid string, err error) {
	entry, err := parser.ParseString(line)
	if err != nil {
		return
	}

	// Save the entry
	for _, curField := range fields {
		switch curField.typ {
		case typ_string:
			strVal, err := entry.Field(curField.name)
			if err != nil {
				return uuid, err
			}
			fmt.Printf("%s is %s\n", curField.name, strVal)
		case typ_float:
			fltVal, err := entry.FloatField(curField.name)
			if err != nil {
				return uuid, err
			}
			fmt.Printf("%s is %f\n", curField.name, fltVal)
		}
	}
	return
}
