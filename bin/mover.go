package main

import (
	"code.google.com/p/go.exp/inotify"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

var (
	source      string
	destination string
)

func deliver(messages chan string) {
	for message := range messages {
		// Our message file name will contain the relative directory and
		// file name. We only care about the file name.
		_, file := path.Split(message)

		// The file will be named "queueName:randomMessageId". Split on the
		// colon to obtain the final target directory.
		pieces := strings.Split(file, ":")

		// Move our source file into the destination using its new name.
		err := os.Rename(path.Join(source, file), path.Join(destination, pieces[0], pieces[1]))

		if err != nil {
			log.Println("Could not move file:", file)
		} else {
			log.Println("Moved file:", file)
		}
	}

	return
}

func main() {
	if len(os.Args) != 4 {
		fmt.Println("mover [source] [destination] [delay]")
		return
	}

	source = os.Args[1]
	destination = os.Args[2]
	delay, err := strconv.Atoi(os.Args[3])

	if err != nil {
		fmt.Println("Delay must be an integer.")
		return
	}

	watcher, err := inotify.NewWatcher()

	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)
	messages := make(chan string)

	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				if delay > 0 {
					go func() {
						time.Sleep(time.Duration(delay) * time.Millisecond)
						messages <- ev.Name
					}()
				} else {
					messages <- ev.Name
				}
			case err := <-watcher.Error:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.AddWatch(source, inotify.IN_CLOSE_WRITE|inotify.IN_MOVED_TO)

	if err != nil {
		log.Fatal(err)
	}

	go deliver(messages)

	<-done

	watcher.Close()
}
