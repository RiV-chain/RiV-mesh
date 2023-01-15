package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hjson/hjson-go"
	"golang.org/x/text/encoding/unicode"

	"github.com/RiV-chain/RiV-mesh/src/defaults"
)

type CmdLineEnv struct {
	args             []string
	endpoint, server string
	injson, ver      bool
}

func newCmdLineEnv() CmdLineEnv {
	var cmdLineEnv CmdLineEnv
	cmdLineEnv.endpoint = defaults.Define().DefaultHttpAddress
	return cmdLineEnv
}

func (cmdLineEnv *CmdLineEnv) parseFlagsAndArgs() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] command [key=value] [key=value] ...\n\n", os.Args[0])
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Please note that options must always specified BEFORE the command\non the command line or they will be ignored.")
		fmt.Println()
		fmt.Println("Commands:\n  - Use \"list\" for a list of available commands")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  - ", os.Args[0], "list")
		fmt.Println("  - ", os.Args[0], "peers")
		fmt.Println("  - ", os.Args[0], "-v self")
		fmt.Println("  - ", os.Args[0], "-endpoint=http://localhost:19019 DHT")
	}

	server := flag.String("endpoint", cmdLineEnv.endpoint, "Admin socket endpoint")
	injson := flag.Bool("json", false, "Output in JSON format (as opposed to pretty-print)")
	ver := flag.Bool("version", false, "Prints the version of this build")

	flag.Parse()

	cmdLineEnv.args = flag.Args()
	cmdLineEnv.server = *server
	cmdLineEnv.injson = *injson
	cmdLineEnv.ver = *ver
}

func (cmdLineEnv *CmdLineEnv) setEndpoint(logger *log.Logger) {
	if cmdLineEnv.server == cmdLineEnv.endpoint {
		if config, err := os.ReadFile(defaults.GetDefaults().DefaultConfigFile); err == nil {
			if bytes.Equal(config[0:2], []byte{0xFF, 0xFE}) ||
				bytes.Equal(config[0:2], []byte{0xFE, 0xFF}) {
				utf := unicode.UTF16(unicode.BigEndian, unicode.UseBOM)
				decoder := utf.NewDecoder()
				config, err = decoder.Bytes(config)
				if err != nil {
					panic(err)
				}
			}
			var dat map[string]interface{}
			if err := hjson.Unmarshal(config, &dat); err != nil {
				panic(err)
			}
			if ep, ok := dat["HttpAddress"].(string); ok && (ep != "none" && ep != "") {
				cmdLineEnv.endpoint = ep
				logger.Println("Found platform default config file", defaults.Define().DefaultHttpAddress)
				logger.Println("Using endpoint", cmdLineEnv.endpoint, "from HttpAddress")
			} else {
				logger.Println("Configuration file doesn't contain appropriate HttpAddress option")
				logger.Println("Falling back to platform default", defaults.Define().DefaultHttpAddress)
			}
		} else {
			logger.Println("Can't open config file from default location", defaults.GetDefaults().DefaultConfigFile)
			logger.Println("Falling back to platform default", defaults.Define().DefaultHttpAddress)
		}
	} else {
		cmdLineEnv.endpoint = cmdLineEnv.server
		logger.Println("Using endpoint", cmdLineEnv.endpoint, "from command line")
	}
}
