package main

import (
	"errors"
	"hash/crc32"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
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
	RootPath      string
	NewFolder     string
	DelayFolder   string
	QueuesFolder  string
	SaveRequests  []chan *SaveRequest
	FetchRequests []chan *FetchRequest
}

func (queue *Queue) Hash() int {
	return int(crc32.ChecksumIEEE([]byte(queue.Id)))
}

func NextFile(dirPath string) string {
	var i int
	var firstFile string

	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		i += 1

		// The walk will include the initial directory as its first "visit."
		// We need to skip that. Returning nil moves us along.
		if i == 1 {
			return nil
		}

		// The walk might bomb out under certain conditions. That's fine.
		if err != nil {
			return err
		}

		firstFile = info.Name()

		return errors.New("Found a file!")
	})

	return firstFile
}

func (store *Store) Prepare(savers int, fetchers int) {
	if store.RootPath == "" {
		panic("No root directory specified!")
	}

	store.NewFolder = path.Join(store.RootPath, "new")
	store.DelayFolder = path.Join(store.RootPath, "delay")
	store.QueuesFolder = path.Join(store.RootPath, "queues")

	os.Mkdir(store.RootPath, 0777)
	os.Mkdir(store.NewFolder, 0777)
	os.Mkdir(store.DelayFolder, 0777)
	os.Mkdir(store.QueuesFolder, 0777)

	store.SaveRequests = make([]chan *SaveRequest, savers)
	store.FetchRequests = make([]chan *FetchRequest, fetchers)

	for i := 0; i < savers; i++ {
		// Make a channel for each go routine.
		store.SaveRequests[i] = make(chan *SaveRequest)

		// Tell the message saver which channel it's supposed to
		// consume from.
		go store.MessageSaver(i)
	}

	for i := 0; i < fetchers; i++ {
		// Make a channel for each go routine.
		store.FetchRequests[i] = make(chan *FetchRequest)

		// Tell the message fetcher which channel it's supposed to
		// consume from.
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

	// Make sure the file is written to disk!
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

func (store *Store) MessageSaver(index int) {
	for {
		request := <-store.SaveRequests[index]
		request.Response <- store.SaveRequestToFile(request)
	}
}

func (store *Store) MessageFetcher(index int) {
	for {
		request := <-store.FetchRequests[index]
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

	// All message save requests need to be serialized on a per-queue
	// basis. There is no race here, but it prevents a single queue from
	// overrunning available i/o.
	store.SaveRequests[queue.Hash()%len(store.SaveRequests)] <- request

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
	store.FetchRequests[queue.Hash()%len(store.FetchRequests)] <- request

	return <-request.Response
}

func (store *Store) DeleteMessage(queue *Queue, message *Message) bool {
	// The currently available messages in the queue is the most likely
	// place the message will exist.
	err := os.Remove(path.Join(store.RootPath, queue.Id, message.Id))

	if err == nil {
		return true
	}

	// Next, the delayed messages folder.
	err = os.Remove(path.Join(store.DelayFolder, queue.Id+":"+message.Id))

	if err == nil {
		return true
	}

	// Finally, the new messages folder.
	err = os.Remove(path.Join(store.NewFolder, queue.Id+":"+message.Id))

	if err == nil {
		return true
	}

	return false
}
