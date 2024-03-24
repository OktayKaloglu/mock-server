package main

import (
	"fmt"
	"log"
	"mock-server/pkg/instance"
	"os"
	"sync"
)

func init() {
	f, err := os.OpenFile("app.log", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	// Set output of logs to the opened file
	log.SetOutput(f)
}

func main() {
	log.Println("Starting application")
	path := "./servers.json"
	var wg sync.WaitGroup
	var mockChannel chan string
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			var s string
			fmt.Scanln(&s)

			mockChannel <- s

			if len(mockChannel) > 2 {

				return
			}
		}
	}()
	defer func() { close(mockChannel) }()
	errs := instance.Run(path, mockChannel, &wg)
	if len(errs) != 0 {
		log.Printf("Errors occurred at mock servers: %v", errs)
	}

	wg.Wait()

}
