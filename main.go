package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/EnemigoPython/go-getit/client"
	"github.com/EnemigoPython/go-getit/runtime"
	"github.com/EnemigoPython/go-getit/server"
)

func main() {
	runTimeFlag := flag.String("runtime", "client", "The runtime mode to execute")
	portFlag := flag.Int("port", 6969, "The port the server will run on")
	storeNameFlag := flag.String("store", "store", "The name of the store file")
	debugFlag := flag.Bool("debug", false, "Run in debug mode")
	flag.Parse()
	config, err := runtime.ParseConfig(
		*runTimeFlag,
		*portFlag,
		*storeNameFlag,
		*debugFlag,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	switch config.RunTime {
	case runtime.Server:
		server.Run()
	case runtime.Client:
		request, err := runtime.ConstructRequest(flag.Args())
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		client.MakeRequest(request)
	}
}
