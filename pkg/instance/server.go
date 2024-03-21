package instance

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

func Run(path string, closeChan <-chan int, wg *sync.WaitGroup) []error {
	servers, err := GetServers(path)
	var errors []error

	if err != nil {
		return append(errors, err)
	}
	for i := range servers {
		errors = append(errors, servers[i].Start(closeChan, wg))
	}

	return errors
}
func (s *Server) Start(closeChan <-chan int, wg *sync.WaitGroup) error {
	// Add 1 to the waitgroup counter for this goroutine
	wg.Add(1)

	go func() error {
		defer wg.Done()

		listener, err := net.Listen("tcp", fmt.Sprintf(":%v", s.Port))
		if err != nil {
			err = fmt.Errorf("Server id: %v,%v", err, s.ServerID)
			log.Print(err)
			return err
		}
		fmt.Printf("Server:%v, is listening at http://127.0.0.1:%v \n", s.ServerID, s.Port)
		for {
			select {
			case id := <-closeChan:
				if s.ServerID == id {

					listener.Close()
					log.Printf("server:%v is closed by channel", s.ServerID)
					return nil
				}
			default:
				// Accept connections and handle them in a separate goroutine
				conn, err := listener.Accept()
				if err != nil {
					err = fmt.Errorf("Server id: %v,%v", err, s.ServerID)
					log.Print(err)
					return err
				}
				go s.handleConnection(conn)
			}
		}
		// Signal that the server has stopped running

	}()
	return nil
}
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Read up to 1024 bytes from connection
	buf := make([]byte, 1024)
	_, err := bufio.NewReader(conn).Read(buf)
	if err != nil {
		fmt.Fprintf(conn, "Error reading request: %v\n", err)
		log.Printf("Server:%v, error while handling request:%v", s.ServerID, err)
		return
	}

	// Convert bytes to string and split by spaces
	request := strings.Split(string(buf), " ")
	if len(request) < 3 { // GET / HTTP/1.1 is at least 3 parts
		log.Printf("Server:%v Invalid request: %v", s.ServerID, request)
		return
	}

	method := strings.ToUpper(request[0]) // Convert to uppercase for switch statement
	fmt.Printf("Server %v, handles the request : %v \n", s.ServerID, request)
	var data []byte
	switch method {
	case "GET":
		// Handle GET request...
		fmt.Print(request[3])
		data, _ = json.Marshal(s.DataMap)

	case "POST":
		// Handle POST request...
		fmt.Fprintf(conn, "Received a POST request\n")
	case "DELETE":
		// Handle DELETE request...
		fmt.Fprintf(conn, "Received a DELETE request\n")
	case "UPDATE":
		// Handle UPDATE request...
		fmt.Fprintf(conn, "Received an UPDATE request\n")
	default:
		// Handle unsupported method...
		fmt.Fprintf(conn, "Unsupported HTTP method: %s\n", method)
	}

	response := "HTTP/1.0 200 OK\r\nContent-Type: application/json; charset=UTF-8\r\nContent-Length: " + strconv.Itoa(len(data)) + "\r\n\r\n" + string(data)
	conn.Write([]byte(response))

}
