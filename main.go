package main

import (
	"gitlab.com/contextualcode/pbrew/cli"
	"gitlab.com/contextualcode/pbrew/core"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
)

func main() {
	// enable output
	output.Enable = true
	// init
	if err := core.InitApp(); err != nil {
		panic(err)
	}
	// execute cli
	if err := cli.Execute(); err != nil {
		panic(err)
	}
}
