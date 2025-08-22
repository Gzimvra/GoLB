package main

import (
	"fmt"
	"net"
)

func main() {
	// Start a TCP listener
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Println("Load balancer listening on :8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// spawn goroutine for each client
		go handleConnection(conn)
	}
}

func handleConnection(c net.Conn) {
	defer c.Close()

	buf := make([]byte, 4096)
	n, _ := c.Read(buf)
	fmt.Println("Received:", string(buf[:n]))

	body := "Hello from the Load Balancer!"
	response := fmt.Sprintf(
		"HTTP/1.1 200 OK\r\nContent-Length: %d\r\nContent-Type: text/plain\r\n\r\n%s",
		len(body), body,
	)

	c.Write([]byte(response))
}
