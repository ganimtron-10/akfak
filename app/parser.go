package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// func parseHeader(buffer []byte) Request {
// 	curHeader := Request{}
// 	curHeader.messageSize = getInt32FromBytes(buffer, 0)
// 	curHeader.apiKey = getInt16FromBytes(buffer, 4)
// 	curHeader.apiVersion = getInt16FromBytes(buffer, 6)
// 	curHeader.correlationId = getInt32FromBytes(buffer, 8)
// 	curHeader.data = buffer[12:]
// 	// fmt.Println(curHeader)
// 	return curHeader
// }

// func decodeHexRequest(hexString string) {
// 	decodedHexString, err := hex.DecodeString(hexString)
// 	fmt.Println(decodedHexString)
// 	if err != nil {
// 		fmt.Println("Error decoding string: ", err.Error())
// 	}
// 	curHeader := Request{}
// 	curHeader.messageSize = getInt32FromBytes(decodedHexString, 0)
// 	curHeader.apiKey = getInt16FromBytes(decodedHexString, 4)
// 	curHeader.apiVersion = getInt16FromBytes(decodedHexString, 6)
// 	curHeader.correlationId = getInt32FromBytes(decodedHexString, 8)
// 	curHeader.data = decodedHexString[12:]
// 	fmt.Println(curHeader)
// }

// func encodeHexRequest(request Request) string {
// 	requestMessage := []byte{}
// 	requestMessage = append(requestMessage, getBytesfromInt16(request.apiKey)...)
// 	requestMessage = append(requestMessage, getBytesfromInt16(request.apiVersion)...)
// 	requestMessage = append(requestMessage, getBytesfromInt32(request.correlationId)...)
// 	requestMessage = append(requestMessage, request.data...)
// 	request.messageSize = calculateMessageSize(requestMessage)

// 	byteRequest := []byte{}
// 	byteRequest = append(byteRequest, getBytesfromInt32(request.messageSize)...)
// 	byteRequest = append(byteRequest, requestMessage...)
// 	fmt.Println(byteRequest)
// 	encodedHexString := hex.EncodeToString(byteRequest)
// 	fmt.Println(encodedHexString)
// 	return encodedHexString
// }

// func util() {
// 	decodeHexRequest("00000023001200046f7fc66100096b61666b612d636c69000a6b61666b612d636c6904302e3100")
// 	// decodeHexRequest("000000230012674a4f74d28b00096b61666b612d636c69000a6b61666b612d636c6904302e3100")
//  000000230012000450b2a73000096b61666b612d636c69000a6b61666b612d636c6904302e3100

// 	request := Request{}
// 	request.apiKey = 18
// 	request.apiVersion = 4
// 	request.correlationId = 1870644833
// 	request.data = []byte("\tkafka-cli\nkafka-cli0.1")
// 	encodeHexRequest(request)
// }

func (request *ApiVersionsRequest) parse(buffer *bytes.Buffer) {
	request.clientSoftwareName = readCompactString(buffer)
	request.clientSoftwareVersion = readCompactString(buffer)
	ignoreTagField(buffer)
}

func parseRequest(buffer *bytes.Buffer) (RequestInterface, error) {
	header := RequestHeader{}
	binary.Read(buffer, binary.BigEndian, &header.messageSize)
	binary.Read(buffer, binary.BigEndian, &header.apiKey)
	binary.Read(buffer, binary.BigEndian, &header.apiVersion)
	binary.Read(buffer, binary.BigEndian, &header.correlationId)
	header.clientId = readNullableString(buffer)
	ignoreTagFiled(buffer)

	switch header.apiKey {
	case 18:
		request := ApiVersionsRequest{RequestHeader: header}
		request.parse(buffer)
		return &request, nil
	default:
		err := fmt.Errorf("%d ApiKey is not Supported", header.apiKey)
		return nil, err
	}

}
