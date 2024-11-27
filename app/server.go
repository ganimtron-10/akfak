package main

import (
	"bytes"
	"encoding/binary"
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
	_, err := connection.Read(buffer)
	if err != nil {
		fmt.Println("Error reading request: ", err.Error())
	}
	// fmt.Println("Recieved Data: ", string(buffer[:n]))

	messageSize := make([]byte, 4)

	correlationID := buffer[8 : 8+4]
	// fmt.Println("CorelationID: ", binary.BigEndian.Uint32(correlationID))
	apiVersion := buffer[6 : 6+2]

	errorCode := make([]byte, 2)
	switch binary.BigEndian.Uint16(apiVersion) {
	case 0, 1, 2, 3, 4:
		binary.BigEndian.PutUint16(errorCode, uint16(0))
	default:
		binary.BigEndian.PutUint16(errorCode, uint16(35))
	}

	apiKeys := []byte{2, 0, 18, 0, 0, 0, 4}

	tagBuffer := byte(0)
	throttleTime := make([]byte, 4)
	binary.BigEndian.PutUint32(throttleTime, uint32(0))
	// buffer = make([]byte, 1024)
	message := correlationID
	message = append(message, errorCode...)
	message = append(message, apiKeys...)
	message = append(message, tagBuffer)
	message = append(message, throttleTime...)
	message = append(message, tagBuffer)

	binary.BigEndian.PutUint32(messageSize, uint32(len(message)))

	bBuffer := bytes.Buffer{}
	bBuffer.Write(messageSize)
	bBuffer.Write(message)
	// Refer for message response structure: https://forum.codecrafters.io/t/question-about-handle-apiversions-requests-stage/1743/3?u=ganimtron-10

	_, err = connection.Write(bBuffer.Bytes())
	// connection.Write([]byte{0, 0, 0, 0, buffer[8], buffer[9], buffer[10], buffer[11], 0, 35})
	if err != nil {
		fmt.Println("Error writing response: ", err.Error())
	}
	// fmt.Println("Send Data: ", string(bBuffer.Bytes()[:n]))

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
