package main

import "fmt"
import "os"

// TODO: Make this time-sequence based so our messages can be at least somewhat
// ordered. Random works for now.
func getRandomUUID() string {
	file, err := os.Open("/dev/urandom")

	if err != nil {
		panic("No random device found!")
	}

	b := make([]byte, 16)

	file.Read(b)
	file.Close()

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
