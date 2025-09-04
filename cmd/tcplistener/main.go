package main

import (
	"fmt"
	"log"
	"net"

	"httpfromtcp/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("starting listener: %+v\n", err)
	}
	defer listener.Close()

	fmt.Println("connection accepted")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("accepting conn: %+v\n", err)
		}
		fmt.Println("accepted connection")

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatalf("error reading request: %+v\n", err)
		}

		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", req.RequestLine.Method)
		fmt.Printf("- Target: %s\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", req.RequestLine.HttpVersion)

		fmt.Println("connection closed")
	}
}
