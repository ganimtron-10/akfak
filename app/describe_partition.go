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

type PartitionRecord struct {
	frameVersion                  uint8
	recordType                    uint8
	version                       uint8
	partitionId                   uuid.UUID
	topicId                       uuid.UUID
	lengthOfReplicaArray          uint64
	replicaArray                  []uint32
	lengthOfInSyncReplicaArray    uint64
	inSyncReplicaArray            []uint32
	lengthOfRemovingReplicasArray uint64
	lengthOfAddingReplicasArray   uint64
	leader                        uint32
	leaderEpoch                   uint32
	partitionEpoch                uint32
	lengthOfDirectoriesArray      uint64
	directoriesArray              []uuid.UUID
	taggedFieldCount              uint64
}

type TopicRecord struct {
	frameVersion     uint8
	recordType       uint8
	version          uint8
	nameLength       uint64
	name             string
	topicId          uuid.UUID
	taggedFieldCount uint64
}

type Record struct {
	length             int64
	attributes         int8
	timestampDelta     int64
	offsetDelta        int64
	keyLength          int64
	key                []byte
	valueLength        int64
	TopicRecord        TopicRecord
	PartitionRecord    PartitionRecord
	FeatureLevelRecord FeatureLevelRecord
	headerArrayCount   uint64
}

type FeatureLevelRecord struct {
	frameVersion     uint8
	recordType       uint8
	version          uint8
	nameLength       uint64
	name             string
	featureLevel     uint16
	taggedFieldCount uint64
}

type ClusterMetadata struct {
	baseOffset           uint64
	batchLength          uint32
	partitionLeaderEpoch uint32
	magicByte            uint8
	crc                  uint32
	attributes           uint16
	lastOffsetDelta      uint32
	baseTimestamp        uint64
	maxTimestamp         uint64
	producerId           uint64
	producerEpoch        uint16
	baseSequence         uint32
	recordsLength        uint32
	records              []Record
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
