package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal("failed to create listener")
	}
	defer listener.Close()
	for {
		reader, err := listener.Accept()
		if err != nil {
			log.Fatal("failed to accept connection")
		}
		fmt.Println("Connection accepted successfully")
		ch := RequestFromReader(reader)
		go printlines(ch)

	}

}

func printlines(ch <-chan string) {
	for line := range ch {
		fmt.Printf("%s\n", line)
	}

}

func getLinesChannel(f io.ReadCloser) <-chan string {
	ch := make(chan string)
	go func() {
		defer f.Close()
		defer close(ch)

		buffer := make([]byte, 8)
		currentline := ""

		for {
			readlen, err := f.Read(buffer)
			if readlen > 0 {
				bytestring := string(buffer[:readlen])
				stringparts := strings.Split(bytestring, "\n")
				currentline = currentline + stringparts[0]
				for i := 1; i < len(stringparts); i++ {
					ch <- currentline
					currentline = stringparts[i]
				}
			}

			// Handle EOF or other errors
			if err != nil {
				// Send the last line even if it doesn't end with a newline
				if currentline != "" {
					ch <- currentline
				}
				fmt.Println("Closing channel")
				break
			}
		}
	}()
	return ch
}
