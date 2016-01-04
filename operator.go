package main

import (
	"flag"
	"github.com/wakeful-deployment/operator/boot"
	"github.com/wakeful-deployment/operator/global"
)

func main() {
	var (
		nodename   = flag.String("node", "", "The name of the host which is running operator")
		consulHost = flag.String("consul", "", "The name or ip of the consul host")
		configPath = flag.String("config", "./operator.json", "The path to the operator.json (default is .)")
		shouldLoop = flag.Bool("loop", false, "Run on each change to the consul key/value storage")
		wait       = flag.String("wait", "5m", "The timeout for polling")
	)
	flag.Parse()

	if *nodename == "" || *consulHost == "" {
		panic("ERROR: Must provide -node and -consul flags")
	}

	global.Info.Nodename = *nodename
	global.Info.Consulhost = *consulHost

	config, err := boot.Boot(*configPath)

	if err != nil {
		panic(err)
	}

	if *shouldLoop {
		boot.Loop(config, *wait)
	} else {
		boot.Once(config)
	}
}
