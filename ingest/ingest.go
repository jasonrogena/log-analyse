package ingest

import (
	"fmt"
	"os"

	"github.com/jasonrogena/log-analyse/config"
	"github.com/jasonrogena/log-analyse/digest"
	"github.com/jasonrogena/log-analyse/types"
)

const Name = "ingest"
const oneOff = "one-off"

type digester interface {
	Absorb(someData interface{}) error
	Digest(someData interface{}) error
	IsDigestable(someData interface{}) bool
}

const typ_string string = "string"
const typ_float string = "float"

var fields = [...]types.FieldType{
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
		digesters, digestErr := getDigesters()
		if configErr == nil && digestErr == nil {
			log := types.Log{Path: args[1], Format: config.Nginx.Format}
			switch ingestType {
			case oneOff:
				ingestOneOff(log, digesters)
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

func getDigesters() (digesters []digester, err error) {
	config, configErr := config.GetConfig()
	if configErr != nil {
		err = configErr
		return
	}

	var urlPathDigInterface digester = digest.InitUrlPathDigester(config.Digest.RbfsLayerCap)
	digesters = append(digesters, urlPathDigInterface)

	return
}

func printHelp() {
	fmt.Printf("Usage %s %s <command>\n\n", os.Args[0], os.Args[1])
	fmt.Printf("Common commands:\n")
	fmt.Printf("\tone-off\t\tIngests the provided file from start to finish\n\n")
	fmt.Printf("\t\t\t\t%s %s one-off <path to log file>\n\n", os.Args[0], os.Args[1])
}
