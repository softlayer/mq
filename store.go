package main

import "os"
import "path"

type Queue struct {
	Id string
}

type Message struct {
	Id      string
	Content string
}

type Store struct {
	Root          string
	NewFolder     string
	CurrentFolder string
	DelayFolder   string
	Folders       [4]string
}

func (store *Store) Prepare() {
	store.Root = "/tmp/mq"

	os.Mkdir(store.Root, 0777)

	store.NewFolder = "new"
	store.CurrentFolder = "current"
	store.DelayFolder = "delay"

	store.Folders = [4]string{
		"",
		store.NewFolder,
		store.CurrentFolder,
		store.DelayFolder,
	}
}

func (store *Store) SaveQueue(queue *Queue) {
	for _, folder := range store.Folders {
		os.Mkdir(path.Join(store.Root, queue.Id, folder), 0777)
	}
}

func (store *Store) LoadQueue(id string) {
	// ...
}

func (store *Store) DeleteQueue(queue *Queue) {
	os.RemoveAll(path.Join(store.Root, queue.Id))
}

func (store *Store) SaveMessage(queue *Queue, message *Message) {
	// ...
}

func (store *Store) LoadMessages(queue *Queue, count int) {
	// ...
}

func (store *Store) DeleteMessage(queue *Queue, id string) {
	// ...
}
