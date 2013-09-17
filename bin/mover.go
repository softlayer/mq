package main

import (
	"flag"
	"log"
	"os"
	"path"
	"strings"
)

var (
	source      string
	destination string
	delay       int
)

func init() {
	flag.StringVar(&source, "source", "/tmp/mq/new", "Source for messages to be moved")
	flag.StringVar(&destination, "destination", "/tmp/mq/queues", "Destination for moved messages")
	flag.IntVar(&delay, "delay", 0, "Delay in milliseconds to wait before moving a message")
	flag.Parse()
}

func main() {
	watch := &Watch{
		Delay:     delay,
		Directory: source,
	}

	watch.Run()

	for file := range watch.Files {
		base := path.Base(file)
		pieces := strings.Split(base, ":")

		err := os.Rename(path.Join(source, base), path.Join(destination, pieces[0], pieces[1]))

		if err != nil {
			log.Print(err)
		}
	}
}
