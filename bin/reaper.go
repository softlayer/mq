package main

import (
	"flag"
	"log"
	"os"
)

var (
	source string
)

func init() {
	flag.StringVar(&source, "source", "/tmp/mq/remove", "Source for messages to be removed")
	flag.Parse()
}

func main() {
	watch := &Watch{
		Directory: source,
	}

	watch.Run()

	for file := range watch.Files {
		err := os.Remove(file)

		if err != nil {
			log.Print(err)
		}
	}
}
