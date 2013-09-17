package main

import (
	"code.google.com/p/go.exp/inotify"
	"log"
	"time"
)

type Watch struct {
	Delay     int
	Directory string
	Files     chan string
}

func (watch *Watch) Run() {
	notify, err := inotify.NewWatcher()

	if err != nil {
		log.Fatal(err)
	}

	watch.Files = make(chan string)

	go func() {
		for {
			select {
			case ev := <-notify.Event:
				if watch.Delay > 0 {
					go func() {
						time.Sleep(time.Duration(watch.Delay) * time.Millisecond)
						watch.Files <- ev.Name
					}()
				} else {
					watch.Files <- ev.Name
				}
			case err := <-notify.Error:
				log.Print(err)
			}
		}
	}()

	err = notify.AddWatch(watch.Directory, inotify.IN_CLOSE_WRITE|inotify.IN_MOVED_TO)

	if err != nil {
		log.Fatal(err)
	} else {
		log.Print(watch)
	}
}
