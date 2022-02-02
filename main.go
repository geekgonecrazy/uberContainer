package main

import (
	"flag"

	"github.com/geekgonecrazy/uberContainer/config"
	"github.com/geekgonecrazy/uberContainer/core"
	"github.com/geekgonecrazy/uberContainer/router"
)

func main() {
	configFile := flag.String("configFile", "config.yaml", "Config File full path. Defaults to current folder")

	flag.Parse()

	if err := config.Load(*configFile); err != nil {
		panic(err)
	}

	config.Load(*configFile)

	core.Init()

	router.Start()
}
