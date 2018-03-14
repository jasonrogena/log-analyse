package ingest

import (
	"fmt"
	"os"

	"github.com/jasonrogena/log-analyse/config"
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
		config, configErr := config.GetConfig()
		if configErr == nil {
			log := Log{path: args[1], format: config.Nginx.Format}
			switch ingestType {
			case oneOff:
				ingestOneOff(log)
			default:
				fmt.Printf("Invalid ingest arguements\n\n")
				printHelp()
			}
		} else {
			fmt.Fprintln(os.Stderr, configErr)
		}
	} else {
		fmt.Printf("Invalid ingest arguements\n\n")
		printHelp()
	}
}

func printHelp() {
	fmt.Printf("Usage %s %s <command>\n\n", os.Args[0], os.Args[1])
	fmt.Printf("Common commands:\n")
	fmt.Printf("\tone-off\t\tIngests the provided file from start to finish\n\n")
	fmt.Printf("\t\t\t\t%s %s one-off <path to log file>\n\n", os.Args[0], os.Args[1])
}
