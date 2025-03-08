package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// ApiVersions

type ApiVersionsRequest struct {
	RequestHeader
	clientSoftwareName    string
	clientSoftwareVersion string
}

type ApiKey struct {
	key        int16
	minVersion int16
	maxVersion int16
}

type ApiVersionsResponse struct {
	errorCode    int16
	numOfApiKeys int8
	apiKeys      []ApiKey
	throttleTime int32
}

func (request *ApiVersionsRequest) parse(buffer *bytes.Buffer) {
	request.clientSoftwareName = readCompactString(buffer)
	request.clientSoftwareVersion = readCompactString(buffer)
	ignoreTagField(buffer)
	fmt.Printf("%+v\n", request)
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
	fmt.Printf("%+v\n", response)
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
