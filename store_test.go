package main

import "os"
import "path"
import "testing"

var root string = "/tmp/mq-test"
var queueId string = "q"
var messageId string = "m"
var messageContent []byte = []byte("c")

func setup() *Store {
	store := &Store{Root: root}
	store.Prepare()

	return store
}

func teardown() {
	os.RemoveAll(root)
}

func TestFolderCreation(t *testing.T) {
	store := setup()

	folders := make(map[string]string)
	folders["root"] = store.Root
	folders["new"] = store.NewFolder
	folders["delay"] = store.DelayFolder
	folders["queues"] = store.QueuesFolder

	for name, folder := range folders {
		if os.Chdir(folder) != nil {
			t.Error("Could not change to", name, "folder at", folder)
		}
	}

	teardown()
}

func TestQueueLifecycle(t *testing.T) {
	store := setup()

	queue := &Queue{Id: queueId}
	queuePath := path.Join(store.QueuesFolder, queue.Id)

	// Can we create a queue?
	store.SaveQueue(queue)

	if os.Chdir(queuePath) != nil {
		t.Error("Could not create queue directory structure")
	}

	// Can we verify a queue exists after creation?
	if store.LoadQueue(queue) == false {
		t.Error("Could not verify queue exists after creation")
	}

	// Can we delete a queue and be certain it no longer exists?
	store.DeleteQueue(queue)

	if os.Chdir(queuePath) == nil {
		t.Error("Our queue directory wasn't properly destroyed")
	}

	teardown()
}

func TestMessageCreation(t *testing.T) {
	store := setup()

	file := queueId + ":" + messageId

	queue := &Queue{Id: queueId}
	message := &Message{Id: messageId, Content: messageContent}
	messagePath := path.Join(store.NewFolder, file)

	store.SaveQueue(queue)
	store.SaveMessage(queue, message)

	stat, err := os.Stat(messagePath)

	if err != nil {
		t.Error("Could not stat message after creation")
	}

	if stat.Name() != file {
		t.Error("Message created with incorrect file name")
	}

	if stat.Size() != int64(len(messageContent)) {
		t.Error("Message created with incorrect file content")
	}

	teardown()
}
