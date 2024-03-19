package main

import (
	"log"
	"mock-server/pkg/kemkum"
)

func main() {
	servers, err := kemkum.GetServers("./servers.json")
	if err != nil {
		log.Fatal(err)
	}
	kemkum.PrintConfigs(servers)
}
