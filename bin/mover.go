package main

import "fmt"
import "syscall"

// import "unsafe"

func main() {
	directory := "inbound"

	fmt.Println("Mover starting up...")

	notify, err := syscall.InotifyInit()

	if err != nil {
		panic("Could not initialize notifier")
	}

	defer syscall.Close(notify)

	watch, err := syscall.InotifyAddWatch(notify, directory, syscall.IN_CREATE|syscall.IN_DELETE)

	if err != nil {
		panic("Could not establish watcher on source directory")
	}

	defer syscall.Close(watch)

	for {
		var buffer [syscall.SizeofInotifyEvent * 4096]byte

		// for {
		n, err := syscall.Read(notify, buffer[:])

		if err != nil {
			fmt.Print(err)
			break
		}

		// Parse the raw event buffer.
		/*
		        for offset <= uint32(n-syscall.SizeofInotifyEvent) {
			    	var offset uint32 = 0

					raw := (*syscall.InotifyEvent)(unsafe.Pointer(&buffer[offset]))
					event := new(Event)
					event.Mask = uint32(raw.Mask)
					event.Cookie = uint32(raw.Cookie)
					nameLen := uint32(raw.Len)
				}
		*/

		fmt.Println(n)
		fmt.Println(buffer)

		// }
	}
}
