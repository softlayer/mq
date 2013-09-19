package main

import (
	"hash/crc32"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
)

type Queue struct {
	Id string
}

type Message struct {
	Id      string
	Content []byte
}

type SaveRequest struct {
	Queue    *Queue
	Message  *Message
	Response chan bool
}

type FetchRequest struct {
	Queue    *Queue
	Response chan *Message
}

type Store struct {
	Workers       int
	Peers         int
	Root          string
	NewFolder     string
	DelayFolder   string
	QueuesFolder  string
	RemoveFolder  string
	SaveRequests  []chan *SaveRequest
	FetchRequests []chan *FetchRequest
}

func Checksum(id string) int {
	return int(crc32.ChecksumIEEE([]byte(id)))
}

func NextFile(dirPath string) string {
	dir, err := os.Open(dirPath)

	if err != nil {
		return ""
	}

	defer dir.Close()

	files, err := dir.Readdir(1)

	if err != nil || len(files) == 0 {
		return ""
	}

	return files[0].Name()
}

func NewStore(workers int, peers int, root string) *Store {
	store := &Store{}
	store.Workers = workers
	store.Peers = peers
	store.Root = root

	return store
}

func (store *Store) PrepareFolders() {
	store.NewFolder = path.Join(store.Root, "new")
	store.DelayFolder = path.Join(store.Root, "delay")
	store.QueuesFolder = path.Join(store.Root, "queues")
	store.RemoveFolder = path.Join(store.Root, "remove")

	os.Mkdir(store.Root, 0777)
	os.Mkdir(store.NewFolder, 0777)
	os.Mkdir(store.DelayFolder, 0777)
	os.Mkdir(store.QueuesFolder, 0777)
	os.Mkdir(store.RemoveFolder, 0777)
}

func (store *Store) PrepareWorkers() {
	store.SaveRequests = make([]chan *SaveRequest, store.Workers)
	store.FetchRequests = make([]chan *FetchRequest, store.Workers)

	for i := 0; i < workers; i++ {
		store.SaveRequests[i] = make(chan *SaveRequest)
		store.FetchRequests[i] = make(chan *FetchRequest)

		go store.MessageSaver(i)
		go store.MessageFetcher(i)
	}
}

func (store *Store) SaveRequestToFile(request *SaveRequest) bool {
	messageFile := request.Queue.Id + ":" + request.Message.Id
	messagePath := path.Join(store.NewFolder, messageFile)

	file, err := os.OpenFile(messagePath, os.O_RDWR|os.O_CREATE, 0777)

	// If we weren't able to open the file for writing,
	// exit early. No need to close it.
	if err != nil {
		return false
	}

	defer file.Close()

	n, err := file.Write(request.Message.Content)

	// Could we write the entire message? If we couldn't, we
	// need to clean up and report back.
	if n < len(request.Message.Content) {
		// Nuke the file...
		os.Remove(messagePath)

		// ...and return a negative response.
		return false
	}

	// If we couldn't write at all, break out. The defer will
	// close the handle.
	if err != nil {
		return false
	}

	file.Sync()

	return true
}

func (store *Store) FetchRequestFromFile(request *FetchRequest) *Message {
	queuePath := path.Join(store.QueuesFolder, request.Queue.Id)
	messageId := NextFile(queuePath)

	if messageId == "" {
		return nil
	}

	messagePath := path.Join(queuePath, messageId)

	// We don't need to lock anything. Our fetch requests have been
	// piplelined to this piont.
	file, err := os.Open(messagePath)

	if err != nil {
		return nil
	}

	defer file.Close()

	// We can still fail here. Make sure to account for a nasty
	// read failure.
	messageContent, err := ioutil.ReadAll(file)

	if err != nil {
		return nil
	}

	message := &Message{
		Id:      messageId,
		Content: messageContent,
	}

	// Got it! Move the file and return our message.
	os.Rename(messagePath, path.Join(store.DelayFolder, request.Queue.Id+":"+message.Id))

	return message
}

func (store *Store) MessageSaver(i int) {
	for {
		request := <-store.SaveRequests[i]
		request.Response <- store.SaveRequestToFile(request)
	}
}

func (store *Store) MessageFetcher(i int) {
	for {
		request := <-store.FetchRequests[i]
		request.Response <- store.FetchRequestFromFile(request)
	}
}

func (store *Store) SaveQueue(queue *Queue) {
	os.Mkdir(path.Join(store.QueuesFolder, queue.Id), 0777)
}

func (store *Store) FetchQueue(queue *Queue) *Queue {
	_, err := os.Stat(path.Join(store.QueuesFolder, queue.Id))

	if err != nil {
		return nil
	}

	// No need to re-allocate, this queue exists. Simply return
	// it to be used.
	return queue
}

func (store *Store) DeleteQueue(queue *Queue) {
	os.RemoveAll(path.Join(store.QueuesFolder, queue.Id))
}

func (store *Store) SaveMessage(queue *Queue, message *Message) bool {
	request := &SaveRequest{
		Queue:    queue,
		Message:  message,
		Response: make(chan bool),
	}

	// It doesn't matter which channel the request is dropped into. What
	// we're concerned about here is this app server clobbering its peers
	// on the same machine.
	store.SaveRequests[rand.Intn(store.Workers)] <- request

	return <-request.Response
}

func (store *Store) FetchMessage(queue *Queue) *Message {
	request := &FetchRequest{
		Queue:    queue,
		Response: make(chan *Message),
	}

	// All message fetch requests need to be serialized on a per-queue
	// basis. This eliminates a "what's the next message in this queue" race
	// on a per-app server basis.
	store.FetchRequests[Checksum(queue.Id)%store.Workers] <- request

	return <-request.Response
}

func (store *Store) DeleteMessage(queue *Queue, message *Message) bool {
	file := queue.Id + ":" + message.Id
	sources := [3]string{
		path.Join(store.QueuesFolder, queue.Id, message.Id),
		path.Join(store.DelayFolder, file),
		path.Join(store.NewFolder, file),
	}

	for _, source := range sources {
		err := os.Rename(source, path.Join(store.RemoveFolder, file))

		if err == nil {
			return true
		}
	}

	return false
}
