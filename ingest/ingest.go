package ingest

import (
	"fmt"
)

const Name = "ingest"
const oneOff = "one-off"

type Log struct {
	path   string
	format string
	uuid   string
}

type Line struct {
	lineNo int64
	value  string
}

type Field struct {
	name string
	typ  string
}

const typ_string string = "string"
const typ_float string = "float"
const ona_access_log_format string = `$http_x_forwarded_for - $remote_user [$time_local]  "$request" $status $body_bytes_sent "$http_referer" "$http_user_agent" $request_time $upstream_response_time $gzip_ratio`

var fields = [...]Field{
	{"http_x_forwarded_for", typ_string},
	{"remote_user", typ_string},
	{"time_local", typ_string},
	{"request", typ_string},
	{"status", typ_float},
	{"body_bytes_sent", typ_float},
	{"http_referer", typ_string},
	{"http_user_agent", typ_string},
	{"request_time", typ_float},
	{"upstream_response_time", typ_float},
	{"gzip_ratio", typ_float},
	{"request_length", typ_float}}

func Process(args []string) {
	if len(args) == 2 {
		ingestType := args[0]
		log := Log{path: args[1], format: ona_access_log_format}
		switch ingestType {
		case oneOff:
			ingestOneOff(log)
		}
	} else {
		printHelp()
	}
}

func printHelp() {
	fmt.Println("Invalid ingest arguements")
}
