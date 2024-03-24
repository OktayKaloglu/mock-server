package instance

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

func Run(path string, closeChan <-chan string, wg *sync.WaitGroup) []error {
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
func (s *Server) Start(closeChan <-chan string, wg *sync.WaitGroup) error {
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
					fmt.Printf("server:%v is closed by channel", s.ServerID)
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

type Response struct {
	Header  string
	Status  string
	Length  string
	Payload string
}

func (r *Response) toString() string {
	return r.Header + r.Status + r.Length + r.Payload
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Read up to 1024 bytes from connection
	buf := make([]byte, 1024)
	_, err := bufio.NewReader(conn).Read(buf)
	if err != nil {
		fmt.Fprintf(conn, "Error reading request: %v\n", err)
		log.Printf("Server:%v, error while handling request err:%v\n", s.ServerID, err)
		return
	}

	// Convert bytes to string and split by spaces
	request := strings.Split(string(buf), " ")
	if len(request) < 3 { // GET / HTTP/1.1 is at least 3 parts
		log.Printf("Server:%v Invalid request: %v\n", s.ServerID, request)
		return
	}

	method := strings.ToUpper(request[0]) // Convert to uppercase for switch statement
	log.Printf("Server %v, handles the request : %v \n", s.ServerID, request)
	endpoints := strings.Split(request[1], "/")[1:] // first "/" is saved as empty string to the array, so we need to remove it
	log.Printf("endpoints array: %v, len:%v \n", endpoints, len(endpoints))
	var data []byte
	switch method {

	case "GET":
		// Handle GET request...
		data, _ = json.Marshal(getKeys(s.DataMap))
		if len(endpoints) > 1 {
			temp, ok := s.DataMap[endpoints[0]]
			if ok {
				temp2, ok2 := temp[endpoints[1]]
				if ok2 {

					data, _ = json.Marshal(temp2)
				} else {
					data, _ = json.Marshal(getKeys(temp))
				}
			}
		} else if len(endpoints) > 0 {
			temp, ok := s.DataMap[endpoints[0]]
			if ok {
				data, _ = json.Marshal(getKeys(temp))
			}
		}

	case "POST":
		// Handle POST request...
		fmt.Printf("Received a POST request\n %v \n", request)
		//409 conflict for same id
		//201 created success and supply the newly created resource with id

		decoder := json.NewDecoder(bytes.NewReader(buf))
		var asd map[string]interface{}
		err := decoder.Decode(&asd)
		if err != nil {
			// Handle decoding error
			fmt.Println(err)
			return
		}
		for key, val := range asd {
			fmt.Printf("key: %v, val:%v\n", key, val)
		}

		//if len(endpoints) > 0 {
		//	a,ok:=s.DataMap[endpoints[0]]
		//	if ok{
		//
		//		a[s.EndPointLen]=
		//	}
		//}

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
func getKeys[T comparable, A any](m map[T]A) []T {
	// Function body to extract keys from the map
	keys := make([]T, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}
