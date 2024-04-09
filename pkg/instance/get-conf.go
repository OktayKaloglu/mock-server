package instance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// todo!
// split in to two different struct, make row struct
// then map the map from the rows
type Server struct {
	stop chan struct{}

	mux         *http.ServeMux
	ServerID    string                   `json:"server_id"`
	IP          string                   `json:"ip"`
	Port        string                   `json:"port"`
	Type        string                   `json:"type"`
	ReturnType  string                   `json:"return_type"`
	Status      string                   `json:"Status"`
	ThreadCount int                      `json:"thread_count"`
	Data        map[string][]interface{} `json:"end_point"`
	EndPointLen int
	DataMap     map[string]map[string]string
}

// Define structs as before (Server and EndPoint)

func ConvertEndPointToMap(data map[string][]interface{}) (map[string]map[string]string, int) {
	result := make(map[string]map[string]string)
	var maxA int = 0
	for key, value := range data {
		innerMap := make(map[string]string)
		for _, item := range value {
			itemMap := item.(map[string]interface{})
			id := fmt.Sprintf("%v", itemMap["id"]) // Access and convert "id" value to string
			itemString := "{"
			for key, val := range itemMap {

				itemString += key + ":" + fmt.Sprintf("%v,", val)
			}
			itemString = itemString[:len(itemString)-1] + "}"
			innerMap[id] = itemString // Use "id" as the key for the item string
		}
		if len(innerMap) > maxA {
			maxA = len(innerMap)
		}
		result[key] = innerMap
	}

	return result, maxA
}
func PrintConfigs(cfg []Server) {

	for _, config := range cfg {
		fmt.Println("Server ID:", config.ServerID)
		fmt.Println("IP:", config.IP)
		fmt.Println("Port:", config.Port)
		fmt.Println("Type:", config.Type)
		fmt.Println("Return Type:", config.ReturnType)
		fmt.Println("Thread Count:", config.ThreadCount)

		fmt.Println("Endpoint Data:")

		for endpointName, endpointData := range config.DataMap {
			fmt.Println("  ", endpointName, ":", endpointData)

		}
	}
}

func GetServers(path string) ([]Server, error) {
	// Read the JSON data from cfg.json
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("Error reading config file:", err)
		return nil, err
	}

	// Create an empty slice to store the parsed configurations
	var cfg []Server

	// Unmarshal the JSON data into the cfg slice
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil, err
	}
	for i := range cfg {
		cfg[i].DataMap, cfg[i].EndPointLen = ConvertEndPointToMap(cfg[i].Data)

	}
	return cfg, nil
}
