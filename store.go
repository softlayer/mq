package main

import "errors"
import "os"
import "io/ioutil"
import "path"
import "path/filepath"

type Queue struct {
	Id string
}

type Message struct {
	Id      string
	Content []byte
}

type Store struct {
	Root         string
	NewFolder    string
	DelayFolder  string
	QueuesFolder string
}

func FirstFileInDir(dirPath string) string {
	var i int = 0
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

func (store *Store) Prepare() {
	store.Root = "/tmp/mq"
	store.NewFolder = path.Join(store.Root, "new")
	store.DelayFolder = path.Join(store.Root, "delay")
	store.QueuesFolder = path.Join(store.Root, "queues")

	os.Mkdir(store.Root, 0777)
	os.Mkdir(store.NewFolder, 0777)
	os.Mkdir(store.DelayFolder, 0777)
	os.Mkdir(store.QueuesFolder, 0777)
}

func (store *Store) SaveQueue(queue *Queue) {
	os.Mkdir(path.Join(store.QueuesFolder, queue.Id), 0777)
}

func (store *Store) LoadQueue(queue *Queue) bool {
	_, err := os.Stat(path.Join(store.Root, queue.Id))

	if err != nil {
		return false
	}

	return true
}

func (store *Store) DeleteQueue(queue *Queue) {
	// TODO: We need to move all files from New and Delay into position to
	// be reaped.
	os.RemoveAll(path.Join(store.Root, queue.Id))
}

func (store *Store) SaveMessage(queue *Queue, message *Message) bool {
	messageFile := queue.Id + ":" + message.Id
	messagePath := path.Join(store.NewFolder, messageFile)

	file, err := os.OpenFile(messagePath, os.O_RDWR|os.O_CREATE, 0777)

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

func (store *Store) LoadNextMessage(queue *Queue) *Message {
	queuePath := path.Join(store.Root, queue.Id)
	messageId := FirstFileInDir(queuePath)

	if messageId == "" {
		return nil
	}

	messagePath := path.Join(queuePath, messageId)

	// TODO: Acquire an exclusive lock on this file before
	// attempting to read.
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

	// Got it! Defer moving the file and return our message.
	defer func() {
		os.Rename(messagePath, path.Join(store.DelayFolder, queue.Id+":"+message.Id))
	}()

	return message
}

func (store *Store) DeleteMessage(queue *Queue, message *Message) bool {
	// The currently available messages in the queue is the most likely
	// place the message will exist.
	err := os.Remove(path.Join(store.Root, queue.Id, message.Id))

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
