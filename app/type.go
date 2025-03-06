package main

import (
	"bytes"

	"github.com/google/uuid"
)

// type Request struct {
// 	messageSize   int32
// 	apiKey        int16
// 	apiVersion    int16
// 	correlationId int32
// 	data          []byte
// }

type RequestHeader struct {
	messageSize   int32
	apiKey        int16
	apiVersion    int16
	correlationId int32
	clientId      string
}

type Response struct {
	messageSize   int32
	correlationId int32

	BytesData bytes.Buffer
}

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

// DescribePartitions

type DescribePartitionsRequest struct {
	RequestHeader
	names                  []string
	responsePartitionLimit int32
	topicName              string
	partitionIndex         int32
}

type Partition struct {
	errorCode      int16
	partitionIndex int32
}

type Topic struct {
	errorCode                 int16
	name                      string
	topicId                   uuid.UUID
	isInternal                bool
	partitions                []Partition
	topicAuthorizedOperations int32
}

type NextCursor struct {
	topicName      string
	partitionIndex int32
}

type DescribePartitionsResponse struct {
	throttleTime int32
	topics       []Topic
	nextCursor   NextCursor
}

type RequestInterface interface {
	parse(buffer *bytes.Buffer)
}

type ResponseInterface interface {
	bytes(buffer *bytes.Buffer, includetagField bool)
}
