package main

import (
	"flag"
	"log"

	"github.com/EnemigoPython/go-getit/src/client"
	"github.com/EnemigoPython/go-getit/src/runtime"
	"github.com/EnemigoPython/go-getit/src/server"
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
		log.Fatal(err)
	}
	switch config.RunTime {
	case runtime.Server:
		server.Run()
	case runtime.Client:
		request, err := runtime.ConstructRequest(flag.Args())
		if err != nil {
			log.Fatal(err)
		}
		client.MakeRequest(request)
	}
}
