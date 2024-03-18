package main

import "github.com/OktayKaloglu/mock-server/instance"

func main() {

	asd, _ := instance.getServers("./config.json")
	instance.printConfigs(asd)
}
