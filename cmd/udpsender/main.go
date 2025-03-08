package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatal("Error resolving address")
	}
	udpConnection, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal("Error dialing up")
	}
	defer udpConnection.Close()
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(">")
		line, err := stdReader.ReadString('\n')
		if err != nil {
			log.Fatal("Error reading stdin")
		}
		_, err = udpConnection.Write([]byte(line))
		if err != nil {
			log.Fatalf("Error writing to udp connection: %v", err)
		}

	}

}
