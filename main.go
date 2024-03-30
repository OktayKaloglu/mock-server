package main

import (
	"context"
	"fmt"
	"log"
	"mock-server/pkg/instance"
	"os"
	"sync"
	"time"
)

func main() {
	f, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	// Set output of logs to the opened file
	log.SetOutput(f)

	log.Println("Starting application")
	path := "./servers.json"
	var wg sync.WaitGroup
	var mockChannel chan string
	defer close(mockChannel)

	servers, errs := instance.Run(path, mockChannel, &wg)
	if len(errs) > 0 {
		for i := range errs {

			log.Printf("Mock server error: %v", errs[i])
		}
	}
	for i := range servers {
		fmt.Printf("sv:%v, port:%v\n", i, servers[i].Addr)
		log.Printf("sv:%v, port:%v\n", i, servers[i].Addr)
	}
	time.Sleep(time.Second * 2)
	for i := range servers {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Set a timeout for graceful shutdown
			defer cancel()

			err := servers[i].Shutdown(ctx)
			if err != nil {
				log.Printf("Error shutting down server1: %v", err)
				fmt.Printf("Error shutting down server1: %v", err)

			} else {
				fmt.Printf("Server:%v closed gracefully\n", i)
				log.Printf("Server:%v closed gracefully\n", i)
				wg.Done()

			}
		}()
	}
	wg.Wait()

}
