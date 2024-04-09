package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"mock-server/pkg/instance"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

func serverGetDataHandle(w http.ResponseWriter, r *http.Request, servers *[]*instance.ServerAndHttp) {

	serverTemplate, err := template.ParseFiles("pkg/web/components/server/serverData.tmpl")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse server template: %v", err), http.StatusInternalServerError)
		return
	}
	urls := cleanURL(r.URL.Path)

	id := ""
	if len(urls) > 1 {
		id = urls[1]
	}

	var jsonData []byte
	if id == "" {

		fileName := "pkg/example/example_server.json" // Replace with your actual file name
		jsonData, err = os.ReadFile(fileName)
		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}
	} else {
		for _, sv := range *servers {
			if sv.ServerStruct.ServerID == id {

				jsonData, err = json.Marshal(sv.ServerStruct)
				if err != nil {
					jsonData = nil
					fmt.Printf("Server:%v, serverData error:%v", id, err)
				}
				break
			}

		}

	}

	err = serverTemplate.Execute(w, struct {
		Server string
	}{string(jsonData)})

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to execute server template: %v", err), http.StatusInternalServerError)
		log.Printf("web server's error: %v\n", err)
		return
	}

}
func serverPostDataHandle(w http.ResponseWriter, r *http.Request, servers *[]*instance.ServerAndHttp, ch *chan string) {

	var server instance.Server

	// Unmarshal the JSON data into the cfg slice
	err := json.Unmarshal([]byte(r.PostFormValue("server")), &server)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create server struct: %v", err), http.StatusInternalServerError)
		log.Printf("create stuck post method error: %v,json:%v\n", err, r.PostFormValue("server"))
		return
	}
	//todo this will update based on the serverID, if post has a different serverID innit it will create new server and add it
	//after addtion we need to change r.url.path to new id
	r.URL.Path = "/serverData/" + server.ServerID
	server.DataMap, server.EndPointLen = instance.ConvertEndPointToMap(server.Data)
	for _, sv := range *servers {
		if sv.ServerStruct.ServerID == server.ServerID {
			if sv.ServerStruct.Status == "available" {
				*ch <- "stop/" + server.ServerID
			}
			sv.ServerStruct = &server
			go func() {
				time.Sleep(time.Second * 10)
				*ch <- "run/" + server.ServerID

			}()
			return
		}
	}

	svvs := instance.ServerAndHttp{ServerHttp: nil, ServerStruct: &server}
	*servers = append(*servers, &svvs)
	*ch <- "run/" + server.ServerID
}

func serverDataHandle(w http.ResponseWriter, r *http.Request, servers *[]*instance.ServerAndHttp, ch *chan string) {
	if r.Method == http.MethodPost {
		serverPostDataHandle(w, r, servers, ch)
	}
	serverGetDataHandle(w, r, servers)

}
func newServersPageHandle(pth string, w http.ResponseWriter, r *http.Request, servers *[]*instance.ServerAndHttp) {
	headerTemplate, err := template.ParseFiles(pth + "/header.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse header template: %v", err), http.StatusInternalServerError)
		return
	}

	serverTemplate, err := template.ParseFiles(pth + "/server.tmpl")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse server template: %v", err), http.StatusInternalServerError)
		return
	}

	footerTemplate, err := template.ParseFiles(pth + "/footer.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse footer template: %v", err), http.StatusInternalServerError)
		return
	}

	// Execute the header, main and footer templates and write them to the response writer
	err = headerTemplate.Execute(w, "header")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to execute header template: %v", err), http.StatusInternalServerError)
		log.Printf("web server's error: %v\n", err)
		return
	}

	urls := cleanURL(r.URL.Path)

	id := ""
	if len(urls) > 1 {
		id = urls[1]
	}

	err = serverTemplate.Execute(w, struct {
		ID string
	}{id})

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to execute server template: %v", err), http.StatusInternalServerError)
		log.Printf("web server's error: %v\n", err)
		return
	}

	err = footerTemplate.Execute(w, "footer")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to execute footer template: %v", err), http.StatusInternalServerError)
		log.Printf("web server's error: %v\n", err)
		return
	}

}

func serversTableHandle(w http.ResponseWriter, r *http.Request, servers *[]*instance.ServerAndHttp) {
	cols := []string{"ServerID", "Port", "Status"}
	var rows []struct {
		ServerID string
		Rows     []string
	}
	for i := range *servers {
		rows = append(rows, struct {
			ServerID string
			Rows     []string
		}{(*servers)[i].ServerStruct.ServerID, []string{(*servers)[i].ServerStruct.ServerID, (*servers)[i].ServerStruct.Port, (*servers)[i].ServerStruct.Status}})
	}

	// Parse the template
	tmpl, err := template.ParseFiles("pkg/web/components/tables/servers.tmpl") // Replace with your template file path
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Execute the template with data
	err = tmpl.Execute(w, struct {
		Cols []string
		Rows []struct {
			ServerID string
			Rows     []string
		}
	}{cols, rows}) // Pass data as a struct
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func serversButtonHandle(w http.ResponseWriter, r *http.Request, servers *instance.Server, ch *chan string, status string) {

	Color := "red"
	Text := "Close"
	URL := "/serversButton/" + servers.ServerID + "/"
	// Parse the template
	tmpl, err := template.ParseFiles("pkg/web/components/buttons/servers.tmpl") // Replace with your template file path
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch {
	case status == "run":
		{
			if servers.Status == "stopped" {
				URL += "red"
				fmt.Println("run/" + servers.ServerID)
				*ch <- "run/" + servers.ServerID

			} else {
				Color = "green"
				Text = "Run"
				URL += "run"

			}

		}

	case status == "red":
		{

			if servers.Status == "available" {
				Color = "green"
				Text = "Run"
				URL += "run"

				*ch <- "stop/" + servers.ServerID
			} else {
				URL += "red"
			}

		}

	case status == "status":
		{
			if servers.Status != "available" {
				Color = "green"
				Text = "Run"
				URL += "run"
			} else {
				URL += "red"
			}
		}
	default:
		{
			http.Error(w, "unknown button status", http.StatusInternalServerError)
		}
	}
	fmt.Printf("%v,%v,%v\n", URL, Color, Text)

	err = tmpl.Execute(w, struct {
		URL   string
		Color string
		Text  string
	}{URL, Color, Text})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func cleanURL(url string) []string {
	rawPths := strings.Split(url, "/")
	var pths []string
	for _, element := range rawPths {
		if !(element == "" || element == " ") {
			pths = append(pths, element)
		}
	}
	return pths
}

// Create a new handler to serve the index.html file with htmx support
func indexHandler(w http.ResponseWriter, r *http.Request, servers *[]*instance.ServerAndHttp, ch *chan string) {
	// Serve the index.html file as a static file
	// Read the header and footer templates
	log.Printf("web server had request: %v\n", r)
	pth := "pkg/web/static"

	fmt.Println(r.URL.Path)
	pths := cleanURL(r.URL.Path)

	switch ln := len(pths); {

	case ln > 0: //this case means request has only one / such as /asd
		{

			switch ep := pths[0]; {
			case ep == "serverData":
				{
					serverDataHandle(w, r, servers, ch)
				}
			case ep == "servers":
				{

					newServersPageHandle(pth, w, r, servers)

				}
			case ep == "shutdown":
				{
					// Create a new template instance and execute it with the header,main, footer, and content templates
					*ch <- "shutdown"

				}
			case ep == "serversButton": //this case means request has two / such as /asd/a
				{
					status := true
					for i := range *servers {
						if pths[1] == (*servers)[i].ServerStruct.ServerID {
							status = false

							serversButtonHandle(w, r, (*servers)[i].ServerStruct, ch, pths[2])
						}
					}
					if status {
						http.Error(w, "Unknown ServerID ", http.StatusInternalServerError)
					}
				}
			case ep == "serversTable":
				{

					serversTableHandle(w, r, servers)
				}
			case ep == "servers":
				{
					status := true
					for i := range *servers {
						if pths[1] == (*servers)[i].ServerStruct.ServerID {
							status = false

							serversButtonHandle(w, r, (*servers)[i].ServerStruct, ch, pths[2])
						}
					}
					if status {
						http.Error(w, "Unknown ServerID ", http.StatusInternalServerError)
					}
				}
			default:
				{
					http.Error(w, "Not supported endpoint", http.StatusInternalServerError)
				}
			}

		}
	case ln == 0: //this case means request is only /
		{

			headerTemplate, err := template.ParseFiles(pth + "/header.html")
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to parse header template: %v", err), http.StatusInternalServerError)
				return
			}

			mainTemplate, err := template.ParseFiles(pth + "/main.html")
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to parse main template: %v", err), http.StatusInternalServerError)
				return
			}

			footerTemplate, err := template.ParseFiles(pth + "/footer.html")
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to parse footer template: %v", err), http.StatusInternalServerError)
				return
			}

			// Execute the header, main and footer templates and write them to the response writer
			err = headerTemplate.Execute(w, "header")
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to execute header template: %v", err), http.StatusInternalServerError)
				log.Printf("web server's error: %v\n", err)
				return
			}

			err = mainTemplate.Execute(w, "main")
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to execute header template: %v", err), http.StatusInternalServerError)
				log.Printf("web server's error: %v\n", err)
				return
			}

			err = footerTemplate.Execute(w, "footer")
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to execute footer template: %v", err), http.StatusInternalServerError)
				log.Printf("web server's error: %v\n", err)
				return
			}

		}
	default:
		http.Error(w, "Not supported path", http.StatusInternalServerError)
	}

}

func Run(wg *sync.WaitGroup, ch *chan string, servers *[]*instance.ServerAndHttp) *http.Server {

	webPort := "7979"
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		indexHandler(w, r, servers, ch)
	})
	sv := &http.Server{Addr: ":" + webPort, Handler: mux}

	fmt.Printf("server: web, Listening on: http://127.0.0.1:%v\n", webPort)

	go func() {

		err := sv.ListenAndServe()
		if err != http.ErrServerClosed {
			fmt.Printf("server:wb, error:%v\n", err)
			log.Printf("Server wb, error: %v", err)
		}

	}()
	wg.Add(1)
	return sv

}
