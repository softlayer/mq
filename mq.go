package main

import "fmt"
import "time"
import "net/http"

type Handler struct {
	router *Router
	endpoints map[string]func()
}

func (handler *Handler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	match := handler.router.Match(req.Method, req.URL.Path)
	
	if match != nil {
		handler.endpoints[match.Name]()
		return
	}
	
	fmt.Println("Not found!")
}

func GetQueue() {
	fmt.Println("Get queue")
}

func CreateQueue() {
	fmt.Println("Create queue")
}

func DeleteQueue() {
	fmt.Println("Delete queue")
}

func main() {
	router := &Router{}
	router.AddRoute("GetQueue", "GET", "^/(?P<queue>[a-z]+)$")
	router.AddRoute("CreateQueue", "PUT", "^/(?P<queue>[a-z]+)$")
	router.AddRoute("DeleteQueue", "DELETE", "^/(?P<queue>[a-z]+)$")

	handler := &Handler{
		router: router,
		endpoints: make(map[string]func()),
	}
	
	// Our handler functions by name. This can easily be looked up by the name
	// our RouteMatch contains.
	handler.endpoints["GetQueue"] = GetQueue
	handler.endpoints["CreateQueue"] = CreateQueue
	handler.endpoints["DeleteQueue"] = DeleteQueue
	
	server := &http.Server{
		Addr: ":8080",
		Handler: handler,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	
	server.ListenAndServe()
}
