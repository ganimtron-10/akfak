package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

type Request struct {
	message_size   int32
	api_key        int16
	api_version    int16
	correlation_id int32
	data           string
}

func getInt32FromBytes(buffer []byte, offset int32) int32 {
	return int32(binary.BigEndian.Uint32(buffer[offset : offset+4+1]))
}

func getInt16FromBytes(buffer []byte, offset int32) int16 {
	return int16(binary.BigEndian.Uint16(buffer[offset : offset+2+1]))
}

func parseRequest(buffer []byte) Request {
	curHeader := Request{}
	curHeader.message_size = getInt32FromBytes(buffer, 0)
	curHeader.api_key = getInt16FromBytes(buffer, 4)
	curHeader.api_version = getInt16FromBytes(buffer, 6)
	curHeader.correlation_id = getInt32FromBytes(buffer, 8)
	curHeader.data = string(buffer[12:])
	// fmt.Println(curHeader)
	return curHeader
}

func decodeHexRequest(hexString string) {
	decodedHexString, err := hex.DecodeString(hexString)
	fmt.Println(decodedHexString)
	if err != nil {
		fmt.Println("Error decoding string: ", err.Error())
	}
	curHeader := Request{}
	curHeader.message_size = getInt32FromBytes(decodedHexString, 0)
	curHeader.api_key = getInt16FromBytes(decodedHexString, 4)
	curHeader.api_version = getInt16FromBytes(decodedHexString, 6)
	curHeader.correlation_id = getInt32FromBytes(decodedHexString, 8)
	curHeader.data = string(decodedHexString[12:])
	fmt.Println(curHeader)
}

func getBytesfromInt32(value int32) []byte {
	byteArray := []byte{}
	byteArray = binary.BigEndian.AppendUint32(byteArray, uint32(value))
	return byteArray
}

func getBytesfromInt16(value int16) []byte {
	byteArray := []byte{}
	byteArray = binary.BigEndian.AppendUint16(byteArray, uint16(value))
	return byteArray
}

func encodeHexRequest(request Request) string {
	requestMessage := []byte{}
	requestMessage = append(requestMessage, getBytesfromInt16(request.api_key)...)
	requestMessage = append(requestMessage, getBytesfromInt16(request.api_version)...)
	requestMessage = append(requestMessage, getBytesfromInt32(request.correlation_id)...)
	requestMessage = append(requestMessage, []byte(request.data)...)
	request.message_size = int32(len(requestMessage))

	byteRequest := []byte{}
	byteRequest = append(byteRequest, getBytesfromInt32(request.message_size)...)
	byteRequest = append(byteRequest, requestMessage...)
	fmt.Println(byteRequest)
	encodedHexString := hex.EncodeToString(byteRequest)
	fmt.Println(encodedHexString)
	return encodedHexString
}

func util() {
	decodeHexRequest("00000023001200046f7fc66100096b61666b612d636c69000a6b61666b612d636c6904302e3100")
	// decodeHexRequest("000000230012674a4f74d28b00096b61666b612d636c69000a6b61666b612d636c6904302e3100")

	request := Request{}
	request.api_key = 18
	request.api_version = 4
	request.correlation_id = 1870644833
	request.data = "\tkafka-cli\nkafka-cli0.1"
	encodeHexRequest(request)
}

// echo "00000023001200046f7fc66100096b61666b612d636c69000a6b61666b612d636c6904302e3100" | xxd -r -p | nc localhost 9092 | hexdump -C
