package main

import "time"
import . "net/http"

type FrontHandler struct {
	Store     *Store
	Router    *Router
	Endpoints map[string]func(*Session)
}

type Session struct {
	Store    *Store
	Match    *RouteMatch
	Request  *Request
	Response ResponseWriter
}

func (handler *FrontHandler) ServeHTTP(response ResponseWriter, request *Request) {
	match := handler.Router.Match(request.Method, request.URL.Path)

	if match != nil {
		session := &Session{
			Store:    handler.Store,
			Match:    match,
			Request:  request,
			Response: response,
		}

		handler.Endpoints[match.Name](session)
		return
	}

	response.WriteHeader(StatusNotFound)
}

func main() {
	store := &Store{Root: "/tmp/mq"}
	store.Prepare()

	router := &Router{}
	router.AddRoute("GetQueue", "GET", "^/(?P<queue>[a-z]+)$")
	router.AddRoute("CreateQueue", "PUT", "^/(?P<queue>[a-z]+)$")
	router.AddRoute("DeleteQueue", "DELETE", "^/(?P<queue>[a-z]+)$")
	router.AddRoute("CreateMessage", "POST", "^/(?P<queue>[a-z]+)/messages$")
	router.AddRoute("GetMessages", "GET", "^/(?P<queue>[a-z]+)/messages$")
	router.AddRoute("DeleteMessage", "DELETE", "^/(?P<queue>[a-z]+)/messages/(?P<message>[a-z]+)$")

	handler := &FrontHandler{
		Store:     store,
		Router:    router,
		Endpoints: make(map[string]func(*Session)),
	}

	// Our handler functions by name. This can easily be looked up by the name
	// our RouteMatch contains.
	handler.Endpoints["GetQueue"] = GetQueue
	handler.Endpoints["CreateQueue"] = CreateQueue
	handler.Endpoints["DeleteQueue"] = DeleteQueue
	handler.Endpoints["CreateMessage"] = CreateMessage
	handler.Endpoints["GetMessages"] = GetMessages
	handler.Endpoints["DeleteMessage"] = DeleteMessage

	server := &Server{
		Addr:           ":8080",
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	server.ListenAndServe()
}
