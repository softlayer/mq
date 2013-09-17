package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

var timeBase = time.Date(1582, time.October, 15, 0, 0, 0, 0, time.UTC).Unix()
var hardwareAddress []byte

func init() {
	interfaces, err := net.Interfaces()

	if err != nil {
		panic("Could not obtain hardware MAC address")
	}

	for _, i := range interfaces {
		if i.Flags&net.FlagLoopback == 0 && len(i.HardwareAddr) > 0 {
			hardwareAddress = i.HardwareAddr
			break
		}
	}
}

// See Christoph Hack's work @ https://github.com/tux21b/gocql/blob/master/uuid/uuid.go
//
// TODO: If we open-source this, we will want to recognize the gocql authors. This is a
// simplified version of their BSD-licensed implementation.
func TimeUUID() string {
	var seq [2]byte
	var uuid [16]byte

	now := time.Now().In(time.UTC)

	// Our time component.
	t := uint64(now.Unix()-timeBase)*10000000 + uint64(now.Nanosecond()/100)
	uuid[0] = byte(t >> 24)
	uuid[1] = byte(t >> 16)
	uuid[2] = byte(t >> 8)
	uuid[3] = byte(t)

	uuid[4] = byte(t >> 40)
	uuid[5] = byte(t >> 32)

	uuid[6] = byte(t>>56) & 0x0F
	uuid[7] = byte(t >> 48)

	// Our random component.
	io.ReadFull(rand.Reader, seq[:])
	uuid[8] = seq[1]
	uuid[9] = seq[0]

	// Our MAC address.
	copy(uuid[10:], hardwareAddress)

	uuid[6] |= 0x10 // Version 1
	uuid[8] &= 0x3F // Clear variant
	uuid[8] |= 0x80 // Set to IETF variant

	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}

func RandomUUID() string {
	file, err := os.Open("/dev/urandom")

	if err != nil {
		panic("No random device found!")
	}

	b := make([]byte, 16)

	file.Read(b)
	file.Close()

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
