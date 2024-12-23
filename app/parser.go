package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/google/uuid"
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

// ApiVersions

func (request *ApiVersionsRequest) parse(buffer *bytes.Buffer) {
	request.clientSoftwareName = readCompactString(buffer)
	request.clientSoftwareVersion = readCompactString(buffer)
	ignoreTagField(buffer)
}

func (response *ApiVersionsResponse) bytes(buffer *bytes.Buffer) {

	binary.Write(buffer, binary.BigEndian, response.errorCode)
	binary.Write(buffer, binary.BigEndian, response.numOfApiKeys)

	for _, apiKey := range response.apiKeys {
		binary.Write(buffer, binary.BigEndian, apiKey.key)
		binary.Write(buffer, binary.BigEndian, apiKey.minVersion)
		binary.Write(buffer, binary.BigEndian, apiKey.maxVersion)
		addTagField(buffer)
	}

	binary.Write(buffer, binary.BigEndian, response.throttleTime)
	addTagField(buffer)
}

func (request *ApiVersionsRequest) generateResponse(commonResponse *Response) {
	commonResponse.correlationId = request.correlationId

	apiVersionResponse := ApiVersionsResponse{}
	apiVersionResponse.errorCode = getApiVersionsErrorCode(request.apiVersion)
	apiVersionResponse.throttleTime = 0

	apiVersion := ApiKey{}
	apiVersion.key = request.apiKey
	apiVersion.minVersion = 0
	apiVersion.maxVersion = 4
	apiVersionResponse.apiKeys = append(apiVersionResponse.apiKeys, apiVersion)

	// describe topic response
	describeTopicVersion := ApiKey{}
	describeTopicVersion.key = 75
	describeTopicVersion.minVersion = 0
	describeTopicVersion.maxVersion = 0
	apiVersionResponse.apiKeys = append(apiVersionResponse.apiKeys, describeTopicVersion)

	apiVersionResponse.numOfApiKeys = int8(len(apiVersionResponse.apiKeys) + 1)

	apiVersionResponse.bytes(&commonResponse.BytesData)
}

// DescribePartitions

func (request *DescribePartitionsRequest) parse(buffer *bytes.Buffer) {
	request.names = getStringArray(buffer)
	binary.Read(buffer, binary.BigEndian, &request.responsePartitionLimit)
	ignoreTagField(buffer)
}

func (response *DescribePartitionsResponse) bytes(buffer *bytes.Buffer) {
	binary.Write(buffer, binary.BigEndian, response.throttleTime)
	binary.Write(buffer, binary.BigEndian, binary.AppendVarint([]byte{}, int64(len(response.topics)+1)))
	for _, topic := range response.topics {
		binary.Write(buffer, binary.BigEndian, topic.errorCode)
		writeCompactString(buffer, topic.name)
		binary.Write(buffer, binary.BigEndian, topic.topicId)
		addTagField(buffer)
	}
	addTagField(buffer)
}

func (request *DescribePartitionsRequest) generateResponse(commonResponse *Response) {
	commonResponse.correlationId = request.correlationId

	dTVResponse := DescribePartitionsResponse{}
	dTVResponse.throttleTime = 0
	dTVResponse.topics = append(dTVResponse.topics, Topic{errorCode: 3, name: request.topicName, topicId: uuid.UUID{0}, partitions: nil})
	dTVResponse.bytes(&commonResponse.BytesData)
}

func (response *Response) bytes(buffer *bytes.Buffer) {
	message := &bytes.Buffer{}
	binary.Write(message, binary.BigEndian, response.correlationId)
	binary.Write(message, binary.BigEndian, response.BytesData.Bytes())
	response.messageSize = int32(message.Len())

	binary.Write(buffer, binary.BigEndian, response.messageSize)
	binary.Write(buffer, binary.BigEndian, message.Bytes())
}

func parseRequest(buffer *bytes.Buffer) (RequestInterface, error) {
	header := RequestHeader{}
	binary.Read(buffer, binary.BigEndian, &header.messageSize)
	binary.Read(buffer, binary.BigEndian, &header.apiKey)
	binary.Read(buffer, binary.BigEndian, &header.apiVersion)
	binary.Read(buffer, binary.BigEndian, &header.correlationId)
	header.clientId = readNullableString(buffer)
	ignoreTagField(buffer)

	switch header.apiKey {
	case 18:
		request := ApiVersionsRequest{RequestHeader: header}
		request.parse(buffer)
		return &request, nil
	case 75:
		request := DescribePartitionsRequest{RequestHeader: header}
		request.parse(buffer)
		return &request, nil
	default:
		err := fmt.Errorf("%d ApiKey is not Supported", header.apiKey)
		return nil, err
	}

}

func processAndGenerateResponse(request RequestInterface) (ResponseInterface, error) {
	switch request := request.(type) {
	case *ApiVersionsRequest:
		response := Response{}
		request.generateResponse(&response)
		return &response, nil
	case *DescribePartitionsRequest:
		response := Response{}
		request.generateResponse(&response)
		return &response, nil
	default:
		err := errors.New("Request type is not Supported")
		return nil, err
	}
}
