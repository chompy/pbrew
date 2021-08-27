package main

import (
	"log"

	"gitlab.com/contextualcode/pbrew/core"
	"gitlab.com/contextualcode/platform_cc/v2/pkg/output"
)

func main() {
	// enable output
	output.Enable = true
	// install brew
	if !core.IsBrewInstalled() {
		done := output.Duration("Installing Homebrew.")
		if err := core.BrewInstall(); err != nil {
			panic(err)
		}
		done()
	}

	proj, err := core.LoadProject(".")
	if err != nil {
		panic(err)
	}
	/*if err := proj.InstallServices(); err != nil {
		panic(err)
	}
	proj.Start()*/

	//log.Println(core.GetHostNames(proj.Routes))

	//log.Println(core.GetRoutesForHostName("www.5dca.org", proj.Routes))

	n, err := proj.GenerateNginxRoutes()
	if err != nil {
		panic(err)
	}
	log.Println(n)

	// load services
	/*services, err := core.LoadServiceList()
	if err != nil {
		panic(err)
	}
	service, err := services.Match("php:7.3")
	if err != nil {
		panic(err)
	}

	if !service.IsInstalled() {
		if err := service.Install(); err != nil {
			panic(err)
		}
	}*/

}
