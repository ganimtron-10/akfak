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
	errorCode     int16
	numOfApiKeys  int8
	apiBytesData  bytes.Buffer
	throttleTime  int32
}

// ApiVersions

type ApiVersionsRequest struct {
	RequestHeader
	clientSoftwareName    string
	clientSoftwareVersion string
}

type ApiVersionsResponse struct {
	apiKey     int16
	minVersion int16
	maxVersion int16
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
	errorCode  int16
	name       string
	topicId    uuid.UUID
	isInternal bool
	partitions []Partition
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
	bytes(buffer *bytes.Buffer)
}
