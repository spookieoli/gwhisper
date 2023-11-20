package webserver

import (
	"fmt"
	"gwhisper_api/whisper_wrapper"
	"html/template"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// WebServer is the main struct for the webserver
type WebServer struct {
	server    *http.Server
	ip        string
	port      int
	templates *template.Template
	Whisper   *whisper_wrapper.Whisper
}

// ConnectRoutes connects the routes to the server
func (ws *WebServer) ConnectRoutes() {
	val := reflect.ValueOf(ws)
	typ := reflect.TypeOf(ws)
	for i := 0; i < val.NumMethod(); i++ {
		method := val.Method(i)
		name := typ.Method(i).Name
		// Skip methods that start with a capital letter
		if name == "ConnectRoutes" || name == "Start" {
			continue
		}
		fmt.Println("Connecting route: " + name)
		// Connect Routes to the server
		if name == "Index" {
			http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				method.Call([]reflect.Value{reflect.ValueOf(w), reflect.ValueOf(r)})
			})
			continue
		}
		http.HandleFunc("/"+strings.ToLower(name), func(w http.ResponseWriter, r *http.Request) {
			method.Call([]reflect.Value{reflect.ValueOf(w), reflect.ValueOf(r)})
		})
	}
}

// Start starts the server
func (ws *WebServer) Start() {
	// Connect the routes
	ws.ConnectRoutes()
	// Start the server
	panic(ws.server.ListenAndServe())
}

// New creates a new WebServer
func New(ip string, port int) *WebServer {
	server := &http.Server{
		Addr:           ip + ":" + strconv.Itoa(port), // IP and port to listen on
		ReadTimeout:    100 * time.Second,             // The maximum duration for reading the entire request, including the body.
		WriteTimeout:   100 * time.Second,             // The maximum duration before timing out writes of the response.
		IdleTimeout:    100 * time.Second,             // The maximum amount of time to wait for the next request when keep-alives are enabled.
		MaxHeaderBytes: 1 << 20,                       // 1 MB max header size means 1 MB max request size
	}
	// Globparse the templates
	templates, err := template.ParseGlob("webserver/html/*.gohtml")
	if err != nil {
		panic(err) // If we can't parse the templates, we can't run the server
	}
	// return the new WebServer
	return &WebServer{server: server, ip: ip, port: port, templates: templates}
}
