package main

import (
	"github.com/geekgonecrazy/uberContainer/config"
	"github.com/geekgonecrazy/uberContainer/core"
	"github.com/geekgonecrazy/uberContainer/router"
)

func main() {

	config.Load("config.yaml")

	core.Init()

	router.Start()
}
