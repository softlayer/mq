package main

import (
	"hash/crc32"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"strings"
)

type Queue struct {
	Id string
}

type Message struct {
	Id      string
	Content []byte
}

type FetchRequest struct {
	Queue    *Queue
	Response chan *Message
}

type Store struct {
	Paths         int
	Workers       int
	RootPaths     []string
	NewFolders    []string
	DelayFolders  []string
	QueuesFolders []string
	RemoveFolders []string
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

func NewStore(workers int, root string) *Store {
	store := &Store{}
	store.RootPaths = strings.Split(root, ",")

	// To be use in the modulo against our checksums for
	// either worker or pathing targets.
	store.Paths = len(store.RootPaths)
	store.Workers = workers

	return store
}

func (store *Store) PrepareFolders() {
	numRoots := len(store.RootPaths)

	if numRoots <= 0 {
		panic("This storage engine requires at least one root path.")
	}

	store.NewFolders = make([]string, numRoots)
	store.DelayFolders = make([]string, numRoots)
	store.QueuesFolders = make([]string, numRoots)
	store.RemoveFolders = make([]string, numRoots)

	for i, root := range store.RootPaths {
		store.NewFolders[i] = path.Join(root, "new")
		store.DelayFolders[i] = path.Join(root, "delay")
		store.QueuesFolders[i] = path.Join(root, "queues")
		store.RemoveFolders[i] = path.Join(root, "remove")
	}

	// As our root paths are all the same underlying file system,
	// we only need to create these directories once.
	os.Mkdir(store.RootPaths[0], 0777)
	os.Mkdir(store.NewFolders[0], 0777)
	os.Mkdir(store.DelayFolders[0], 0777)
	os.Mkdir(store.QueuesFolders[0], 0777)
	os.Mkdir(store.RemoveFolders[0], 0777)
}

func (store *Store) PrepareWorkers() {
	store.FetchRequests = make([]chan *FetchRequest, store.Workers)

	for i := 0; i < workers; i++ {
		store.FetchRequests[i] = make(chan *FetchRequest)

		go store.MessageFetcher(i)
	}
}

func (store *Store) FetchRequestFromFile(request *FetchRequest) *Message {
	queuePath := path.Join(store.QueuesFolders[rand.Intn(store.Paths)], request.Queue.Id)
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
	os.Rename(messagePath, path.Join(store.DelayFolders[rand.Intn(store.Paths)], request.Queue.Id+":"+message.Id))

	return message
}

func (store *Store) MessageFetcher(index int) {
	for {
		request := <-store.FetchRequests[index]
		request.Response <- store.FetchRequestFromFile(request)
	}
}

func (store *Store) SaveQueue(queue *Queue) {
	os.Mkdir(path.Join(store.QueuesFolders[rand.Intn(store.Paths)], queue.Id), 0777)
}

func (store *Store) FetchQueue(queue *Queue) *Queue {
	_, err := os.Stat(path.Join(store.QueuesFolders[rand.Intn(store.Paths)], queue.Id))

	if err != nil {
		return nil
	}

	// No need to re-allocate, this queue exists. Simply return
	// it to be used.
	return queue
}

func (store *Store) DeleteQueue(queue *Queue) {
	os.RemoveAll(path.Join(store.QueuesFolders[rand.Intn(store.Paths)], queue.Id))
}

func (store *Store) SaveMessage(queue *Queue, message *Message) bool {
	messageFile := queue.Id + ":" + message.Id
	messagePath := path.Join(store.NewFolders[rand.Intn(store.Paths)], messageFile)

	file, err := os.OpenFile(messagePath, os.O_RDWR|os.O_CREATE, 0777)

	// If we weren't able to open the file for writing,
	// exit early. No need to close it.
	if err != nil {
		return false
	}

	defer file.Close()

	n, err := file.Write(message.Content)

	// Could we write the entire message? If we couldn't, we
	// need to clean up and report back.
	if n < len(message.Content) {
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

	return true
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
	// The currently available messages in the queue is the most likely
	// place the message will exist.
	err := os.Remove(path.Join(store.QueuesFolders[rand.Intn(store.Paths)], queue.Id, message.Id))

	if err == nil {
		return true
	}

	// Next, the delayed messages folder.
	err = os.Remove(path.Join(store.DelayFolders[rand.Intn(store.Paths)], queue.Id+":"+message.Id))

	if err == nil {
		return true
	}

	// Finally, the new messages folder.
	err = os.Remove(path.Join(store.NewFolders[rand.Intn(store.Paths)], queue.Id+":"+message.Id))

	if err == nil {
		return true
	}

	return false
}
