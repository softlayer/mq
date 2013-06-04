package main

import (
	"io/ioutil"
	"net/http"
)

// TODO: Every call requires the "queue" parameter. Maybe move this into a
// middleware-like mechanism to allocate a queue struct if present?

func GetQueue(session *Session) {
	queue := &Queue{Id: session.Match.Variables["queue"]}

	if session.Store.LoadQueue(queue) {
		session.Response.WriteHeader(http.StatusOK)
		return
	}

	session.Response.WriteHeader(http.StatusNotFound)
}

func CreateQueue(session *Session) {
	queue := &Queue{Id: session.Match.Variables["queue"]}

	session.Store.SaveQueue(queue)
	session.Response.WriteHeader(http.StatusCreated)
}

func DeleteQueue(session *Session) {
	queue := &Queue{Id: session.Match.Variables["queue"]}

	session.Store.DeleteQueue(queue)
	session.Response.WriteHeader(http.StatusOK)
}

func GetMessage(session *Session) {
	queue := &Queue{Id: session.Match.Variables["queue"]}

	message := session.Store.LoadNextMessage(queue)

	if message != nil {
		session.Response.Header().Set("X-Message-Id", message.Id)
		session.Response.WriteHeader(http.StatusOK)
		session.Response.Write(message.Content)
		return
	}

	session.Response.WriteHeader(http.StatusNoContent)
}

func CreateMessage(session *Session) {
	queue := &Queue{Id: session.Match.Variables["queue"]}

	body, err := ioutil.ReadAll(session.Request.Body)

	if err != nil {
		session.Response.WriteHeader(http.StatusBadRequest)
		return
	}

	message := &Message{
		Id:      getRandomUUID(),
		Content: body,
	}

	// TODO: Check to see if this exploded before sending back
	// a successful status.
	session.Store.SaveMessage(queue, message)
	session.Response.Header().Set("X-Message-Id", message.Id)
	session.Response.WriteHeader(http.StatusCreated)
}

func DeleteMessage(session *Session) {
	queue := &Queue{Id: session.Match.Variables["queue"]}
	message := &Message{Id: session.Match.Variables["message"]}

	session.Store.DeleteMessage(queue, message)
	session.Response.WriteHeader(http.StatusAccepted)
}
