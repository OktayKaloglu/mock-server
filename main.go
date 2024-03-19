package main

import (
	"log"

	_ "github.com/OktayKaloglu/mock-server/kemkum" // Assuming kemkum is a private package, replace with the correct import path if it's public.
)

func main() {
	servers, err := kemkum.GetServers("../config.json")
	if err != nil {
		log.Fatal(err) // handle error properly by using log.Fatal instead of plain print statement.
	}

	kemkum.PrintConfigs(servers)
}
