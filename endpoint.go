package main

import (
	"io/ioutil"
	"net/http"
)

func GetQueue(session *Session) {
	queue := &Queue{Id: session.Match.Variables["queue"]}

	if session.Store.FetchQueue(queue) != nil {
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

	message := session.Store.FetchMessage(queue)

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
		Id:      TimeUUID(),
		Content: body,
	}

	success := session.Store.SaveMessage(queue, message)

	if success {
		session.Response.Header().Set("X-Message-Id", message.Id)
		session.Response.WriteHeader(http.StatusCreated)
		return
	}

	session.Response.Header().Set("Retry-After", "10")
	session.Response.WriteHeader(http.StatusServiceUnavailable)
}

func DeleteMessage(session *Session) {
	queue := &Queue{Id: session.Match.Variables["queue"]}
	message := &Message{Id: session.Match.Variables["message"]}

	session.Store.DeleteMessage(queue, message)
	session.Response.WriteHeader(http.StatusAccepted)
}
