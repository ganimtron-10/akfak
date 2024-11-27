package main

import (
	"fmt"
	"net"
	"os"
)

// Ensures gofmt doesn't remove the "net" and "os" imports in stage 1 (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func handleConnection(connection net.Conn) {
	defer connection.Close()

	buffer := make([]byte, 1024)
	n, err := connection.Read(buffer)
	if err != nil {
		fmt.Println("Error reading request: ", err.Error())
	}
	fmt.Println("Recieved Data: ", string(buffer[:n]))

	// buffer = make([]byte, 1024)
	buffer = []byte("1\r\n \r\n7")
	n, err = connection.Write(buffer)
	if err != nil {
		fmt.Println("Error writing response: ", err.Error())
	}
	fmt.Println("Send Data: ", string(buffer[:n]))

}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage

	l, err := net.Listen("tcp", "0.0.0.0:9092")
	if err != nil {
		fmt.Println("Failed to bind to port 9092")
		os.Exit(1)
	}
	defer l.Close()

	for {
		connection, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handleConnection(connection)
	}
}
