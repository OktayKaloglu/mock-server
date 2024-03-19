package main

import (
	"log"

	_ "github.com/OktayKaloglu/mock-server/kemkum/kemkum"
)

func main() {
	servers, err := kemkum.GetServers("../config.json")
	if err != nil {
		log.Fatal(err) // handle error properly by using log.Fatal instead of plain print statement.
	}

	kemkum.PrintConfigs(servers)
}
