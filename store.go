package main

import "os"
import "path"

type Queue struct {
	Id string
}

type Message struct {
	Id      string
	Content []byte
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

func (store *Store) LoadQueue(queue *Queue) {
	// ...
}

func (store *Store) DeleteQueue(queue *Queue) {
	os.RemoveAll(path.Join(store.Root, queue.Id))
}

func (store *Store) SaveMessage(queue *Queue, message *Message) bool {
	path := path.Join(store.Root, queue.Id, store.NewFolder, message.Id)

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0777)

	// If we weren't able to open the file for writing,
	// exit early. No need to close it.
	if err != nil {
		return false
	}

	defer file.Close()

	// TODO: Check to make sure how much we wrote is how
	// much we were given.
	_, err = file.Write(message.Content)

	// If we couldn't write, break out. The defer will
	// close the handle.
	if err != nil {
		return false
	}

	// Make sure the file is written to disk!
	file.Sync()

	return true
}

func (store *Store) LoadMessages(queue *Queue, count int) {
	// ...
}

func (store *Store) DeleteMessage(queue *Queue, message *Message) {
	// ...
}
