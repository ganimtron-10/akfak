package main

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/google/uuid"
)

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

func (request *DescribePartitionsRequest) parse(buffer *bytes.Buffer) {
	request.names = getStringArray(buffer)
	binary.Read(buffer, binary.BigEndian, &request.responsePartitionLimit)
	ignoreTagField(buffer)
	fmt.Printf("%+v\n", request)
}

func (response *DescribePartitionsResponse) bytes(buffer *bytes.Buffer) {

	binary.Write(buffer, binary.BigEndian, response.throttleTime)

	// topics
	binary.Write(buffer, binary.BigEndian, int8(len(response.topics)+1))
	for _, topic := range response.topics {
		binary.Write(buffer, binary.BigEndian, topic.errorCode)
		writeCompactString(buffer, topic.name)
		// buffer.WriteString(topic.topicId)
		binary.Write(buffer, binary.BigEndian, topic.topicId[:])
		binary.Write(buffer, binary.BigEndian, topic.isInternal)

		// partitions
		if topic.partitions == nil {
			binary.Write(buffer, binary.BigEndian, int8(1))
		} else {
			binary.Write(buffer, binary.BigEndian, int8(len(topic.partitions)+1))
			for _, partition := range topic.partitions {
				binary.Write(buffer, binary.BigEndian, partition.errorCode)
				binary.Write(buffer, binary.BigEndian, partition.partitionIndex)
				addTagField(buffer)
			}
		}

		binary.Write(buffer, binary.BigEndian, topic.topicAuthorizedOperations)
		addTagField(buffer)
	}

	// next cursor
	// binary.Write(buffer, binary.BigEndian, int8(1))
	// writeCompactString(buffer, response.nextCursor.topicName)
	// binary.Write(buffer, binary.BigEndian, response.nextCursor.partitionIndex)
	// null cursor
	binary.Write(buffer, binary.BigEndian, int8(-1))

	addTagField(buffer)
	fmt.Printf("%+v\n", response)
}

func (request *DescribePartitionsRequest) generateResponse(commonResponse *Response) {
	commonResponse.correlationId = request.correlationId

	dTVResponse := DescribePartitionsResponse{}
	dTVResponse.throttleTime = 0
	dTVResponse.topics = append(dTVResponse.topics, Topic{errorCode: 0, name: request.names[0], topicId: uuid.UUID{0}, partitions: nil})
	dTVResponse.bytes(&commonResponse.BytesData)
}
