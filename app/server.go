package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
)

func handleConnection(connection net.Conn) {
	defer connection.Close()

	buffer := make([]byte, 1024)
	_, err := connection.Read(buffer)
	if err != nil {
		fmt.Println("Error reading request: ", err.Error())
		return
	}

	messageSize := int32(4) // 4 bytes for the correlation_id
	correlationID := int32(7)

	// Prepare response buffer
	response := make([]byte, 8)
	binary.BigEndian.PutUint32(response[0:4], uint32(messageSize))        // message_size
	binary.BigEndian.PutUint32(response[4:8], uint32(correlationID)) // correlation_id

	_, err = connection.Write(response)
	if err != nil {
		fmt.Println("Error writing response: ", err.Error())
	}
}

func main() {
	fmt.Println("Logs from your program will appear here!")

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
