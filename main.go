package main

import (
	"gitlab.com/contextualcode/pbrew/cli"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
)

func main() {
	// enable output
	output.Enable = true
	// execute cli
	if err := cli.Execute(); err != nil {
		panic(err)
	}
}
