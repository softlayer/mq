package main

import (
	"os"
	"path"
	"testing"
)

var (
	testRoot       string = "/tmp/mq-test"
	testWorkers    int    = 1
	queueId        string = "q"
	messageId      string = "m"
	messageContent []byte = []byte("abcdefghijklmnopqrstuvwxyz")
)

func setup() *Store {
	store := NewStore(testWorkers, testRoot)

	store.PrepareFolders()
	store.PrepareWorkers()

	return store
}

func teardown() {
	os.RemoveAll(testRoot)
}

func TestFolderCreation(t *testing.T) {
	store := setup()

	folders := make(map[string]string)
	folders["root"] = store.RootPaths[0]
	folders["new"] = store.NewFolders[0]
	folders["delay"] = store.DelayFolders[0]
	folders["queues"] = store.QueuesFolders[0]

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
	queuePath := path.Join(store.QueuesFolders[0], queue.Id)

	// Can we create a queue?
	store.SaveQueue(queue)

	if os.Chdir(queuePath) != nil {
		t.Error("Could not create queue directory structure")
	}

	// Can we verify a queue exists after creation?
	if store.FetchQueue(queue) == nil {
		t.Error("Could not verify queue exists after creation")
	}

	// Can we delete a queue and be certain it no longer exists?
	store.DeleteQueue(queue)

	if os.Chdir(queuePath) == nil {
		t.Error("Queue directory wasn't properly destroyed")
	}

	teardown()
}

func TestMessageLifecycle(t *testing.T) {
	store := setup()

	file := queueId + ":" + messageId

	queue := &Queue{Id: queueId}
	message := &Message{Id: messageId, Content: messageContent}

	// All the pathing we will need to check as the message moves through its
	// lifecycle.
	messagePathNew := path.Join(store.NewFolders[0], file)
	messagePathAvailable := path.Join(store.QueuesFolders[0], queueId, messageId)
	messagePathDelay := path.Join(store.DelayFolders[0], file)

	store.SaveQueue(queue)
	store.SaveMessage(queue, message)

	stat, err := os.Stat(messagePathNew)

	// Make sure the file was created with the correct name and content.
	if err != nil {
		t.Error("Could not stat message after creation")
	}

	if stat.Name() != file {
		t.Error("Message created with incorrect file name")
	}

	if stat.Size() != int64(len(messageContent)) {
		t.Error("Message created with incorrect file content")
	}

	// Note: we are simulating an external daemon. Normally the storage
	// mechanism has no idea how to sort messages.
	os.Rename(messagePathNew, messagePathAvailable)

	// When we fetch a message, it should contain the same ID we saved to the
	// new folder.
	if store.FetchMessage(queue).Id != message.Id {
		t.Error("Unable to fetch correct message")
	}

	// Verify after we fetched the message was temporarily moved to the delay
	// folder.
	stat, err = os.Stat(messagePathDelay)

	if err != nil {
		t.Error("Could not stat message after fetching")
	}

	if stat.Name() != file {
		t.Error("Message moved with incorrect file name")
	}

	if stat.Size() != int64(len(messageContent)) {
		t.Error("Message moved with incorrect file content")
	}

	// Finally, delete the message.
	if store.DeleteMessage(queue, message) == false {
		t.Error("Could not delete message")
	}

	// We will attempt to delete the message from all the places it could
	// exist. We should always get a non-nil response (an error) from each of
	// these calls.
	if os.Remove(messagePathNew) == nil {
		t.Error("Found message file in 'new'")
	}

	if os.Remove(messagePathAvailable) == nil {
		t.Error("Found message file in 'available'")
	}

	if os.Remove(messagePathDelay) == nil {
		t.Error("Found message file in 'delay'")
	}

	teardown()
}

func BenchmarkMessageCreation(b *testing.B) {
	store := setup()

	queue := &Queue{Id: queueId}
	store.SaveQueue(queue)

	for i := 0; i < b.N; i++ {
		message := &Message{Id: getRandomUUID(), Content: messageContent}
		store.SaveMessage(queue, message)
	}

	teardown()
}
