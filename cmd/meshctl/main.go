package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/RiV-chain/RiV-mesh/src/version"
)

func main() {
	// makes sure we can use defer and still return an error code to the OS
	os.Exit(run())
}

func run() int {
	logbuffer := &bytes.Buffer{}
	logger := log.New(logbuffer, "", log.Flags())

	defer func() int {
		if r := recover(); r != nil {
			logger.Println("Fatal error:", r)
			fmt.Print(logbuffer)
			return 1
		}
		return 0
	}()

	cmdLineEnv := newCmdLineEnv()
	cmdLineEnv.parseFlagsAndArgs()

	if cmdLineEnv.ver {
		fmt.Println("Build name:", version.BuildName())
		fmt.Println("Build version:", version.BuildVersion())
		fmt.Println("To get the version number of the running Mesh node, run", os.Args[0], "getSelf")
		return 0
	}

	if len(cmdLineEnv.args) == 0 {
		flag.Usage()
		return 0
	}

	cmdLineEnv.setEndpoint(logger)

	u, err := url.Parse(cmdLineEnv.endpoint)

	if err == nil {
		var response *http.Response
		var err error
		if cmdLineEnv.injson {
			response, err = http.Get(u.String() + "/api/" + cmdLineEnv.args[0])
		} else {
			response, err = http.Get(u.String() + "/api/" + cmdLineEnv.args[0] + "?fmt=table")
		}
		if err != nil {
			panic(err)
		}
		result, err := io.ReadAll(response.Body)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(result))
	} else {
		panic(err)
	}

	return 0
}
