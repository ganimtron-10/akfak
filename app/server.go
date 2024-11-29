package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
)

func handleConnection(connection net.Conn) {

	buffer := make([]byte, 1024)
	_, err := connection.Read(buffer)
	if err != nil {
		fmt.Println("Error reading request: ", err.Error())
	}
	// fmt.Println("Recieved Data: ", string(buffer[:n]))

	// ----------- New Method -------------

	// request := parseRequest(buffer)
	// fmt.Println(request)

	// ----------- Old Method -------------
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
	if err != nil {
		fmt.Println("Error writing response: ", err.Error())
	}
	// fmt.Println("Send Data: ", string(bBuffer.Bytes()[:n]))

}

func main() {
	PORT := 9092
	fmt.Printf("Starting Akfak on port %d...", PORT)

	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", PORT))
	if err != nil {
		fmt.Printf("Failed to bind to port %d", PORT)
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
