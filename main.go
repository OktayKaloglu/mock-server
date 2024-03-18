package main

import "github.com/OktayKaloglu/mock-server/kemkum"

func main() {

	asd, _ := kemkum.getServers("./config.json")
	kemkum.printConfigs(asd)
}
