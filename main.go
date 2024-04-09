package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mock-server/pkg/instance"
	"mock-server/pkg/web"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

func listenChannel(wg *sync.WaitGroup, ch *chan string, servers *[]*instance.ServerAndHttp, wb *http.Server) {
	defer wg.Done()
	for { // Infinite loop
		message := <-*ch // Receive data from the channel and store it in 'message'
		fmt.Println("Received message:", message)

		method := strings.Split(message, "/")
		if message == "shutdown" {

			for _, server := range *servers {
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Set a timeout for graceful shutdown
					defer cancel()
					err := server.ServerHttp.Shutdown(ctx)
					if err != nil {
						log.Printf("Error shutting down server1: %v", err)
						fmt.Printf("Error shutting down server1: %v", err)
					} else {
						fmt.Printf("Server:%v closed gracefully\n", server.ServerStruct.ServerID)
						log.Printf("Server:%v closed gracefully\n", server.ServerStruct.ServerID)
						wg.Done()
						server.ServerStruct.Status = "stopped"
					}
				}()
			}
			if wb != nil {

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Set a timeout for graceful shutdown
				defer cancel()
				err := wb.Shutdown(ctx)
				if err != nil {
					log.Printf("Error shutting down server1: %v", err)
					fmt.Printf("Error shutting down server1: %v", err)
				} else {
					fmt.Printf("Server:wb closed gracefully\n")
					log.Printf("Server:wb closed gracefully\n")
					wg.Done()

				}
			}
			wg.Done()
			return
		}

		for _, server := range *servers {
			if method[1] == server.ServerStruct.ServerID {
				if method[0] == "stop" {
					go func() {
						ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Set a timeout for graceful shutdown
						defer cancel()
						err := server.ServerHttp.Shutdown(ctx)
						if err != nil {
							log.Printf("Error shutting down server1: %v", err)
							fmt.Printf("Error shutting down server1: %v", err)
						} else {
							fmt.Printf("Server:%v closed gracefully\n", server.ServerStruct.ServerID)
							log.Printf("Server:%v closed gracefully\n", server.ServerStruct.ServerID)
							wg.Done()
							server.ServerStruct.Status = "stopped"
						}
					}()
				} else if method[0] == "run" {
					sv, err := server.ServerStruct.Start(wg)
					if err != nil {
						log.Printf("error starting server:%v,error:%v\n", server.ServerStruct.ServerID, err)
						fmt.Printf("error starting server:%v,error:%v\n", server.ServerStruct.ServerID, err)
						continue
					}
					server.ServerHttp = sv
					server.ServerStruct.Status = "available"

				}
			}

		}
	}
}
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
	mockChannel := make(chan string)
	defer close(mockChannel)
	var errs []error
	var servers []*instance.ServerAndHttp

	//run the web server
	wg.Add(1)

	sv := web.Run(&wg, &mockChannel, &servers)

	wg.Add(1)
	go func() {
		listenChannel(&wg, &mockChannel, &servers, sv)
	}()
	//wait for the web server to get the web socket before any mock server
	time.Sleep(time.Second * 2)

	//start the mock servers
	servers, errs = instance.Run(path, mockChannel, &wg)
	// if any error is accrued at mock server innitilatizion log them
	if len(errs) > 0 {
		for i := range errs {

			log.Printf("Mock server error: %v", errs[i])
		}
	}
	//output the mock servers addresses
	for i := range servers {
		fmt.Printf("sv:%v, port:%v\n", i, servers[i].ServerHttp.Addr)
		log.Printf("sv:%v, port:%v\n", i, servers[i].ServerHttp.Addr)
	}
	wg.Wait()

	var jsonData []byte
	var svss []*instance.Server
	for _, server := range servers {
		svss = append(svss, server.ServerStruct)
	}
	jsonData, err = json.Marshal(svss)

	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	// Write the JSON data to a file
	err = os.WriteFile("user.json", jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing JSON to file:", err)
		return
	}

	fmt.Println("User data saved to user.json successfully!")

}
