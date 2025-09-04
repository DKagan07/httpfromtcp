package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	udp, err := net.ResolveUDPAddr("udp", ":42069")
	if err != nil {
		log.Fatalf("failed resolving udp: %+v\n", err)
	}

	udpConn, err := net.DialUDP("udp", nil, udp)
	if err != nil {
		log.Fatalf("creating udp connection: %+v\n", err)
	}
	defer udpConn.Close()

	buf := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf(">")
		string, err := buf.ReadString('\n')
		if err != nil {
			log.Fatalf("failed to read string from buffer: %+v\n", err)
		}

		if _, err := udpConn.Write([]byte(string)); err != nil {
			log.Fatalf("failed to write to udpConn: %+v\n", err)
		}
	}
}
