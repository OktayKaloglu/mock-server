package instance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
)

func Run(path string, closeChan <-chan string, wg *sync.WaitGroup) ([]*http.Server, []error) {
	servers, err := GetServers(path)
	var errors []error
	if err != nil {
		errors = append(errors, err)
		return nil, errors
	}

	var sv []*http.Server
	for i := range servers {
		sv1, err1 := servers[i].Start(wg)
		if err1 != http.ErrServerClosed {
			sv = append(sv, sv1)
		} else {
			errors = append(errors, err1)
		}

	}

	return sv, errors
}
func (s *Server) Start(wg *sync.WaitGroup) (*http.Server, error) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", s.Handler)
	sv := &http.Server{Addr: ":" + s.Port, Handler: mux}

	fmt.Printf("server: %v , Listening on: http://127.0.0.1:%v\n", s.ServerID, s.Port)
	var err error
	wg.Add(1)

	go func() {
		err = sv.ListenAndServe()
		if err != http.ErrServerClosed {
			fmt.Printf("server:%v error:%v\n", s.ServerID, err)
			log.Printf("Server %v, error: %v", s.ServerID, err)

		}
	}()

	return sv, err
}
func getKeys[T comparable, A any](m map[T]A) []T {
	// Function body to extract keys from the map
	keys := make([]T, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

// todo!
// make every method separate function
type Response struct {
	Status  int
	Message string
	Data    interface{}
}

func (s *Server) Handler(w http.ResponseWriter, r *http.Request) {

	response := Response{}
	endpoints := strings.Split(r.URL.Path, "/")[1:] //split the path for "/" char, first "/" is added to list as empty string
	log.Printf("Server %v, handles the request : %v \n", s.ServerID, r)
	switch r.Method {
	case "GET":
		response.Data = fmt.Sprint(getKeys(s.DataMap))
		response.Message = "Available endpoints are provided"
		response.Status = http.StatusBadRequest
		if len(endpoints) > 0 {
			temp, ok := s.DataMap[endpoints[0]]
			response.Message = "Available entities are provided"
			if ok {
				response.Data = fmt.Sprint(getKeys(temp))
				if len(endpoints) > 1 {
					temp2, ok2 := temp[endpoints[1]]
					if ok2 {
						response.Status = http.StatusAccepted
						response.Message = "Requested entity is provided"
						response.Data = temp2
					}
				}
			}
		}

	case "POST":
		// implementation for POST method
		body := new(bytes.Buffer)
		_, err := io.Copy(body, r.Body)
		if err != nil {
			response.Message = "Error: Failed to read body"
			response.Status = http.StatusInternalServerError
		} else {
			defer r.Body.Close()
			jsonData := make(map[string]interface{})
			err = json.Unmarshal([]byte(body.Bytes()), &jsonData)
			if err != nil {
				response.Message = "Error: Failed to unmarshal JSON data"
				response.Status = http.StatusBadRequest
			} else {

				response.Data = fmt.Sprint(getKeys(s.DataMap))
				response.Message = "Available endpoints are provided"
				response.Status = http.StatusBadRequest
				//implement, if response has a id with it, add to map with that supplied id
				if len(endpoints) > 0 {
					temp, ok := s.DataMap[endpoints[0]]
					if ok {
						s.EndPointLen += 1
						response.Data = nil
						response.Status = http.StatusAccepted
						response.Message = "POST is successful, ID: " + fmt.Sprint(s.EndPointLen)
						temp[fmt.Sprint(s.EndPointLen)] = convertMapJson(&jsonData)

					}
				}

			}
		}

	case "DELETE":
		response.Data = fmt.Sprint(getKeys(s.DataMap))
		response.Message = "Available endpoints are provided"
		response.Status = http.StatusBadRequest
		if len(endpoints) > 0 {
			temp, ok := s.DataMap[endpoints[0]]
			response.Message = "Available entities are provided"
			if ok {
				response.Data = fmt.Sprint(getKeys(temp))
				if len(endpoints) > 1 {
					_, ok2 := temp[endpoints[1]]
					if ok2 {
						response.Message = "Requested entity is deleted"
						s.EndPointLen -= 1
						delete(temp, endpoints[1])
					}
				}
			}
		}

	case "PUT":
		// implementation for POST method
		body := new(bytes.Buffer)
		_, err := io.Copy(body, r.Body)
		if err != nil {
			response.Message = "Error: Failed to read body"
			response.Status = http.StatusInternalServerError
		} else {
			defer r.Body.Close()
			jsonData := make(map[string]interface{})
			err = json.Unmarshal([]byte(body.Bytes()), &jsonData)
			if err != nil {
				response.Message = "Error: Failed to unmarshal JSON data"
				response.Status = http.StatusBadRequest
			} else {

				response.Data = fmt.Sprint(getKeys(s.DataMap))
				response.Message = "Available endpoints are provided"
				response.Status = http.StatusBadRequest
				//implement, if response has a id with it, add to map with that supplied id
				if len(endpoints) > 0 {
					temp, ok := s.DataMap[endpoints[0]]
					if ok {
						response.Data = fmt.Sprint(getKeys(temp))
						response.Message = "Available entities are provided"
						_, ok2 := temp[endpoints[1]]
						if ok2 {
							temp[endpoints[1]] = convertMapJson(&jsonData)
							response.Data = nil
							response.Status = http.StatusAccepted
							response.Message = "PUT is successful, ID: " + fmt.Sprint(s.EndPointLen)

						}

					}
				}

			}
		}

	default:
		response.Message = "Unsupported method"
		response.Status = http.StatusMethodNotAllowed

	}
	//todo!
	//responded data: is not json object its string like "{a:1,b:2}"
	json.NewEncoder(w).Encode(response)
	log.Printf("server:%v ,response: %v\n", s.ServerID, fmt.Sprint(response))
}
func convertMapJson(mp *map[string]interface{}) string {
	s := "{"
	for key, val := range *mp {

		s += fmt.Sprintf("%v:%v,", key, val)
	}
	s = s[:len(s)-1] + "}"
	return s
}
