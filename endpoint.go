package main

import "net/http"

func GetQueue(session *Session) {
	session.Response.WriteHeader(http.StatusOK)
}

func CreateQueue(session *Session) {
	queue := &Queue{
		Id: session.Match.Variables["queue"],
	}

	session.Store.SaveQueue(queue)
	session.Response.WriteHeader(http.StatusCreated)
}

func DeleteQueue(session *Session) {
	queue := &Queue{
		Id: session.Match.Variables["queue"],
	}

	session.Store.DeleteQueue(queue)
	session.Response.WriteHeader(http.StatusOK)
}

func GetMessages(session *Session) {
	// ...
}

func CreateMessage(session *Session) {
	// ...
}

func DeleteMessage(session *Session) {
	// ...
}
