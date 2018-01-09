package ingest

import (
	"fmt"
)

func writeLine(log Log, line string) (uuid string, err error) {
	fmt.Println(line)
	return
}
