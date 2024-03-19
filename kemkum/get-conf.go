package kemkum

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {

	fmt.Println("asd")

}

// Define structs as before (Server and EndPoint)
type Server struct {
	ServerID    int                      `json:"server_id"`
	IP          string                   `json:"ip"`
	Port        string                   `json:"port"`
	Type        string                   `json:"type"`
	ReturnType  string                   `json:"return_type"`
	ThreadCount int                      `json:"thread_count"`
	Data        map[string][]interface{} `json:"end_point"`
	DataMap     map[string]map[string]string
}

func convertEndPointToMap(data map[string][]interface{}) map[string]map[string]string {
	result := make(map[string]map[string]string)

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
		result[key] = innerMap
	}

	return result
}
func printConfigs(cfg []Server) {

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

func getServers(path string) ([]Server, error) {
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
		cfg[i].DataMap = convertEndPointToMap(cfg[i].Data)
	}
	return cfg, nil
}
