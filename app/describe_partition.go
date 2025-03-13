package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/google/uuid"
)

type DescribePartitionsRequest struct {
	RequestHeader
	names                  []string
	responsePartitionLimit int32
	topicName              string
	partitionIndex         int32
}

type Partition struct {
	errorCode              int16
	partitionIndex         int32
	leaderId               int32
	leaderEpoch            int32
	replicaNodes           []int32
	isrNodes               []int32
	eligibleLeaderReplicas []int32
	lastKnownElr           []int32
	offlineReplicas        []int32
}

type Topic struct {
	errorCode                 int16
	name                      string
	topicId                   uuid.UUID
	isInternal                bool
	partitions                []*Partition
	topicAuthorizedOperations int32
}

type NextCursor struct {
	topicName      string
	partitionIndex int32
}

type DescribePartitionsResponse struct {
	throttleTime int32
	topics       []*Topic
	nextCursor   NextCursor
}

type PartitionRecord struct {
	frameVersion                  uint8
	recordType                    uint8
	version                       uint8
	partitionId                   int32
	topicId                       uuid.UUID
	lengthOfReplicaArray          uint64
	replicaArray                  []int32
	lengthOfInSyncReplicaArray    uint64
	inSyncReplicaArray            []int32
	lengthOfRemovingReplicasArray uint64
	lengthOfAddingReplicasArray   uint64
	leader                        int32
	leaderEpoch                   int32
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
	frameVersion       uint8
	recordType         uint8
	version            uint8
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
	records              []*Record
}

func (request *DescribePartitionsRequest) parse(buffer *bytes.Buffer) {
	request.names = getStringArray(buffer)
	binary.Read(buffer, binary.BigEndian, &request.responsePartitionLimit)
	ignoreTagField(buffer)
	fmt.Printf("%+v\n", request)
}

func (response *DescribePartitionsResponse) bytes(buffer *bytes.Buffer, request *DescribePartitionsRequest) {

	binary.Write(buffer, binary.BigEndian, response.throttleTime)

	topicsToSend := []*Topic{}

	for _, topic := range response.topics {
		if slices.Contains(request.names, topic.name) {
			topicsToSend = append(topicsToSend, topic)
		}
	}

	binary.Write(buffer, binary.BigEndian, int8(len(topicsToSend)+1))
	for _, topic := range topicsToSend {
		binary.Write(buffer, binary.BigEndian, topic.errorCode)
		writeCompactString(buffer, topic.name)

		binary.Write(buffer, binary.BigEndian, topic.topicId[:])
		binary.Write(buffer, binary.BigEndian, topic.isInternal)

		fmt.Printf("\nPartition Record in bytes: %+v", topic.partitions)

		if topic.partitions == nil {
			binary.Write(buffer, binary.BigEndian, int8(1))
		} else {
			binary.Write(buffer, binary.BigEndian, int8(len(topic.partitions)+1))
			for _, partition := range topic.partitions {
				binary.Write(buffer, binary.BigEndian, partition.errorCode)
				binary.Write(buffer, binary.BigEndian, partition.partitionIndex)
				binary.Write(buffer, binary.BigEndian, partition.leaderId)
				binary.Write(buffer, binary.BigEndian, partition.leaderEpoch)
				writeCompactArray(buffer, partition.replicaNodes)
				writeCompactArray(buffer, partition.isrNodes)
				writeCompactArray(buffer, partition.eligibleLeaderReplicas)
				writeCompactArray(buffer, partition.lastKnownElr)
				writeCompactArray(buffer, partition.offlineReplicas)
				addTagField(buffer)
			}
		}

		binary.Write(buffer, binary.BigEndian, topic.topicAuthorizedOperations)
		addTagField(buffer)
	}

	binary.Write(buffer, binary.BigEndian, int8(-1))

	addTagField(buffer)
	fmt.Printf("%+v\n", response)
}

func (request *DescribePartitionsRequest) generateResponse(commonResponse *Response) {
	commonResponse.correlationId = request.correlationId

	dTVResponse := DescribePartitionsResponse{}
	dTVResponse.throttleTime = 0
	// dTVResponse.topics = append(dTVResponse.topics, Topic{errorCode: 0, name: request.names[0], topicId: uuid.UUID{0}, partitions: nil})

	clusterMetadataLogs, err := readClusterMetadata()
	if err != nil {
		fmt.Printf("Error while reading cluster data. Error details: %s", err)
	}

	err = addClusterMetadataIntoResponse(&dTVResponse, clusterMetadataLogs)
	if err != nil {
		fmt.Printf("Error while adding cluster data into repsonse. Error details: %s", err)
	}

	dTVResponse.bytes(&commonResponse.BytesData, request)

}

func addClusterMetadataIntoResponse(response *DescribePartitionsResponse, clusterMetadataLogs []*ClusterMetadata) error {

	topicPartitionMap := make(map[uuid.UUID]*Topic)

	for _, clusterMetadata := range clusterMetadataLogs {
		// fmt.Printf("%+v\n%d\n", clusterMetadata.records[0], clusterMetadata.recordsLength)

		for _, record := range clusterMetadata.records {
			switch record.recordType {
			case 2:
				// topicRecord
				topic := &Topic{
					errorCode:  0,
					name:       record.TopicRecord.name,
					topicId:    record.TopicRecord.topicId,
					isInternal: false,
				}

				topicPartitionMap[topic.topicId] = topic
				response.topics = append(response.topics, topic)
			case 3:
				// partitionRecord
				partition := &Partition{
					errorCode:              0,
					partitionIndex:         record.PartitionRecord.partitionId,
					leaderId:               record.PartitionRecord.leader,
					leaderEpoch:            record.PartitionRecord.leaderEpoch,
					replicaNodes:           record.PartitionRecord.replicaArray,
					isrNodes:               record.PartitionRecord.inSyncReplicaArray,
					eligibleLeaderReplicas: []int32{},
					lastKnownElr:           []int32{},
					offlineReplicas:        []int32{},
				}

				topic, ok := topicPartitionMap[record.PartitionRecord.topicId]
				if ok {
					topic.partitions = append(topic.partitions, partition)
				}

				fmt.Printf("\nTopic details: %+v", topic)
				fmt.Printf("\nPartition Record adding in response: %+v", record.PartitionRecord)
			case 12:
				// featureRecord
				// Skipping for now
			}
		}
	}

	return nil
}

func readClusterMetadata() ([]*ClusterMetadata, error) {

	clusterMetadataLogFileName := "/tmp/kraft-combined-logs/__cluster_metadata-0/00000000000000000000.log"
	fileData, err := os.ReadFile(clusterMetadataLogFileName)
	if err != nil {
		fmt.Printf("Error while reading cluster metadata log file, Error Details: %s", err)
	}
	// fmt.Printf("%+v\n", fileData)

	// fileData := []byte{0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 79, 0, 0, 0, 1, 2, 176, 105, 69, 124, 0, 0, 0, 0, 0, 0, 0, 0, 1, 145, 224, 90, 248, 24, 0, 0, 1, 145, 224, 90, 248, 24, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 0, 0, 0, 1, 58, 0, 0, 0, 1, 46, 1, 12, 0, 17, 109, 101, 116, 97, 100, 97, 116, 97, 46, 118, 101, 114, 115, 105, 111, 110, 0, 20, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 154, 0, 0, 0, 1, 2, 105, 208, 150, 103, 0, 0, 0, 0, 0, 1, 0, 0, 1, 145, 224, 91, 45, 21, 0, 0, 1, 145, 224, 91, 45, 21, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 0, 0, 0, 2, 60, 0, 0, 0, 1, 48, 1, 2, 0, 4, 98, 97, 114, 0, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 84, 0, 0, 144, 1, 0, 0, 2, 1, 130, 1, 1, 3, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 84, 2, 0, 0, 0, 1, 2, 0, 0, 0, 1, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 2, 16, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 154, 0, 0, 0, 1, 2, 16, 140, 191, 92, 0, 0, 0, 0, 0, 1, 0, 0, 1, 145, 224, 91, 45, 21, 0, 0, 1, 145, 224, 91, 45, 21, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 0, 0, 0, 2, 60, 0, 0, 0, 1, 48, 1, 2, 0, 4, 98, 97, 122, 0, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 152, 0, 0, 144, 1, 0, 0, 2, 1, 130, 1, 1, 3, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 152, 2, 0, 0, 0, 1, 2, 0, 0, 0, 1, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 2, 16, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 6, 0, 0, 0, 228, 0, 0, 0, 1, 2, 133, 202, 140, 77, 0, 0, 0, 0, 0, 2, 0, 0, 1, 145, 224, 91, 45, 21, 0, 0, 1, 145, 224, 91, 45, 21, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 0, 0, 0, 3, 60, 0, 0, 0, 1, 48, 1, 2, 0, 4, 112, 97, 120, 0, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 99, 0, 0, 144, 1, 0, 0, 2, 1, 130, 1, 1, 3, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 99, 2, 0, 0, 0, 1, 2, 0, 0, 0, 1, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 2, 16, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 1, 0, 0, 144, 1, 0, 0, 4, 1, 130, 1, 1, 3, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 99, 2, 0, 0, 0, 1, 2, 0, 0, 0, 1, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 2, 16, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 1, 0, 0}

	// fileData := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 79, 0, 0, 0, 1, 2, 176, 105, 69, 124, 0, 0, 0, 0, 0, 0, 0, 0, 1, 145, 224, 90, 248, 24, 0, 0, 1, 145, 224, 90, 248, 24, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 0, 0, 0, 1, 58, 0, 0, 0, 1, 46, 1, 12, 0, 17, 109, 101, 116, 97, 100, 97, 116, 97, 46, 118, 101, 114, 115, 105, 111, 110, 0, 20, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 228, 0, 0, 0, 1, 2, 36, 219, 18, 221, 0, 0, 0, 0, 0, 2, 0, 0, 1, 145, 224, 91, 45, 21, 0, 0, 1, 145, 224, 91, 45, 21, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 0, 0, 0, 3, 60, 0, 0, 0, 1, 48, 1, 2, 0, 4, 115, 97, 122, 0, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 145, 0, 0, 144, 1, 0, 0, 2, 1, 130, 1, 1, 3, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 145, 2, 0, 0, 0, 1, 2, 0, 0, 0, 1, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 2, 16, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 1, 0, 0, 144, 1, 0, 0, 4, 1, 130, 1, 1, 3, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 145, 2, 0, 0, 0, 1, 2, 0, 0, 0, 1, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 2, 16, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 1, 0, 0}

	// fileData := []byte{0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 228, 0, 0, 0, 1, 2, 36, 219, 18, 221, 0, 0, 0, 0, 0, 2, 0, 0, 1, 145, 224, 91, 45, 21, 0, 0, 1, 145, 224, 91, 45, 21, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 0, 0, 0, 3, 60, 0, 0, 0, 1, 48, 1, 2, 0, 4, 115, 97, 122, 0, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 145, 0, 0, 144, 1, 0, 0, 2, 1, 130, 1, 1, 3, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 145, 2, 0, 0, 0, 1, 2, 0, 0, 0, 1, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 2, 16, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 1, 0, 0, 144, 1, 0, 0, 4, 1, 130, 1, 1, 3, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 145, 2, 0, 0, 0, 1, 2, 0, 0, 0, 1, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 2, 16, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 1, 0, 0}

	// fileData := []byte{0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 79, 0, 0, 0, 1, 2, 176, 105, 69, 124, 0, 0, 0, 0, 0, 0, 0, 0, 1, 145, 224, 90, 248, 24, 0, 0, 1, 145, 224, 90, 248, 24, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 0, 0, 0, 1, 58, 0, 0, 0, 1, 46, 1, 12, 0, 17, 109, 101, 116, 97, 100, 97, 116, 97, 46, 118, 101, 114, 115, 105, 111, 110, 0, 20, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 0, 0, 0, 154, 0, 0, 0, 1, 2, 52, 167, 186, 136, 0, 0, 0, 0, 0, 1, 0, 0, 1, 145, 224, 91, 45, 21, 0, 0, 1, 145, 224, 91, 45, 21, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 0, 0, 0, 2, 60, 0, 0, 0, 1, 48, 1, 2, 0, 4, 98, 97, 122, 0, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 83, 0, 0, 144, 1, 0, 0, 2, 1, 130, 1, 1, 3, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 83, 2, 0, 0, 0, 1, 2, 0, 0, 0, 1, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 2, 16, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 0, 154, 0, 0, 0, 1, 2, 34, 245, 28, 91, 0, 0, 0, 0, 0, 1, 0, 0, 1, 145, 224, 91, 45, 21, 0, 0, 1, 145, 224, 91, 45, 21, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 0, 0, 0, 2, 60, 0, 0, 0, 1, 48, 1, 2, 0, 4, 102, 111, 111, 0, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 33, 0, 0, 144, 1, 0, 0, 2, 1, 130, 1, 1, 3, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 33, 2, 0, 0, 0, 1, 2, 0, 0, 0, 1, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 2, 16, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 6, 0, 0, 0, 228, 0, 0, 0, 1, 2, 73, 229, 226, 168, 0, 0, 0, 0, 0, 2, 0, 0, 1, 145, 224, 91, 45, 21, 0, 0, 1, 145, 224, 91, 45, 21, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 0, 0, 0, 3, 60, 0, 0, 0, 1, 48, 1, 2, 0, 4, 112, 97, 120, 0, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 133, 0, 0, 144, 1, 0, 0, 2, 1, 130, 1, 1, 3, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 133, 2, 0, 0, 0, 1, 2, 0, 0, 0, 1, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 2, 16, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 1, 0, 0, 144, 1, 0, 0, 4, 1, 130, 1, 1, 3, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 133, 2, 0, 0, 0, 1, 2, 0, 0, 0, 1, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 2, 16, 0, 0, 0, 0, 0, 64, 0, 128, 0, 0, 0, 0, 0, 0, 1, 0, 0}

	// clusterMetadata, err := parseClusterMetadata(bytes.NewBuffer(fileData))
	// if err != nil {
	// 	return &ClusterMetadata{}, err
	// }
	// // _ = clusterMetadata
	// fmt.Printf("%+v\n", clusterMetadata)

	clusterMetadataLogRecords := []*ClusterMetadata{}
	fileBuffer := bytes.NewBuffer(fileData)

	for fileBuffer.Len() > 0 {
		clusterMetadata, err := parseClusterMetadata(fileBuffer)
		if err != nil {
			return []*ClusterMetadata{}, err
		}
		// fmt.Printf("%+v\n", clusterMetadata)
		clusterMetadataLogRecords = append(clusterMetadataLogRecords, clusterMetadata)
	}

	return clusterMetadataLogRecords, nil
}

func parseClusterMetadata(fileBuffer *bytes.Buffer) (*ClusterMetadata, error) {
	clusterMetadata := &ClusterMetadata{}

	binary.Read(fileBuffer, binary.BigEndian, &clusterMetadata.baseOffset)
	binary.Read(fileBuffer, binary.BigEndian, &clusterMetadata.batchLength)
	binary.Read(fileBuffer, binary.BigEndian, &clusterMetadata.partitionLeaderEpoch)
	binary.Read(fileBuffer, binary.BigEndian, &clusterMetadata.magicByte)
	binary.Read(fileBuffer, binary.BigEndian, &clusterMetadata.crc)
	binary.Read(fileBuffer, binary.BigEndian, &clusterMetadata.attributes)
	binary.Read(fileBuffer, binary.BigEndian, &clusterMetadata.lastOffsetDelta)
	binary.Read(fileBuffer, binary.BigEndian, &clusterMetadata.baseTimestamp)
	binary.Read(fileBuffer, binary.BigEndian, &clusterMetadata.maxTimestamp)
	binary.Read(fileBuffer, binary.BigEndian, &clusterMetadata.producerId)
	binary.Read(fileBuffer, binary.BigEndian, &clusterMetadata.producerEpoch)
	binary.Read(fileBuffer, binary.BigEndian, &clusterMetadata.baseSequence)
	binary.Read(fileBuffer, binary.BigEndian, &clusterMetadata.recordsLength)

	// fmt.Printf("ClusterMetadata Records Length: %d\n", clusterMetadata.recordsLength)

	for i := uint32(0); i < clusterMetadata.recordsLength; i++ {
		record := Record{}

		length, _ := binary.ReadVarint(fileBuffer)
		record.length = length
		_ = binary.Read(fileBuffer, binary.BigEndian, &record.attributes)
		timestampDelta, _ := binary.ReadVarint(fileBuffer)
		record.timestampDelta = timestampDelta
		offsetDelta, _ := binary.ReadVarint(fileBuffer)
		record.offsetDelta = offsetDelta
		keyLength, _ := binary.ReadVarint(fileBuffer)
		record.keyLength = keyLength
		if keyLength > 0 {
			record.key = make([]byte, keyLength)
			_, _ = io.ReadFull(fileBuffer, record.key)
		}
		valueLength, _ := binary.ReadVarint(fileBuffer)
		record.valueLength = valueLength
		valueBytes := make([]byte, valueLength)
		_, _ = io.ReadFull(fileBuffer, valueBytes)
		valueBuf := bytes.NewBuffer(valueBytes)
		_ = binary.Read(valueBuf, binary.BigEndian, &record.frameVersion)
		_ = binary.Read(valueBuf, binary.BigEndian, &record.recordType)
		_ = binary.Read(valueBuf, binary.BigEndian, &record.version)
		switch record.recordType {
		case 2:
			topicRecord := TopicRecord{}
			topicRecord.frameVersion = record.frameVersion
			topicRecord.recordType = record.recordType
			topicRecord.version = record.version
			topicRecord.nameLength, _ = binary.ReadUvarint(valueBuf)
			topicRecord.name = string(valueBuf.Next(int(topicRecord.nameLength - 1)))
			topicIDBytes := valueBuf.Next(16)
			topicRecord.topicId, _ = uuid.FromBytes(topicIDBytes)
			topicRecord.taggedFieldCount, _ = binary.ReadUvarint(valueBuf)
			record.TopicRecord = topicRecord
		case 3:
			partitionRecord := PartitionRecord{}
			partitionRecord.frameVersion = record.frameVersion
			partitionRecord.recordType = record.recordType
			partitionRecord.version = record.version
			_ = binary.Read(valueBuf, binary.BigEndian, &partitionRecord.partitionId)
			_ = binary.Read(valueBuf, binary.BigEndian, &partitionRecord.topicId)
			partitionRecord.lengthOfReplicaArray, _ = binary.ReadUvarint(valueBuf)
			partitionRecord.replicaArray = make([]int32, partitionRecord.lengthOfReplicaArray)
			for j := uint64(0); j < partitionRecord.lengthOfReplicaArray; j++ {
				_ = binary.Read(valueBuf, binary.BigEndian, &partitionRecord.replicaArray[j])
			}
			partitionRecord.lengthOfInSyncReplicaArray, _ = binary.ReadUvarint(valueBuf)
			partitionRecord.inSyncReplicaArray = make([]int32, partitionRecord.lengthOfInSyncReplicaArray)
			for j := uint64(0); j < partitionRecord.lengthOfInSyncReplicaArray; j++ {
				_ = binary.Read(valueBuf, binary.BigEndian, &partitionRecord.inSyncReplicaArray[j])
			}
			partitionRecord.lengthOfRemovingReplicasArray, _ = binary.ReadUvarint(valueBuf)
			partitionRecord.lengthOfAddingReplicasArray, _ = binary.ReadUvarint(valueBuf)
			_ = binary.Read(valueBuf, binary.BigEndian, &partitionRecord.leader)
			_ = binary.Read(valueBuf, binary.BigEndian, &partitionRecord.leaderEpoch)
			_ = binary.Read(valueBuf, binary.BigEndian, &partitionRecord.partitionEpoch)
			partitionRecord.lengthOfDirectoriesArray, _ = binary.ReadUvarint(valueBuf)
			partitionRecord.directoriesArray = make([]uuid.UUID, partitionRecord.lengthOfDirectoriesArray)
			for j := uint64(0); j < partitionRecord.lengthOfDirectoriesArray; j++ {
				dirBytes := valueBuf.Next(16)
				partitionRecord.directoriesArray[j], _ = uuid.FromBytes(dirBytes)
			}
			partitionRecord.taggedFieldCount, _ = binary.ReadUvarint(valueBuf)
			record.PartitionRecord = partitionRecord

			fmt.Printf("\nPartition Record: %+v", partitionRecord)
		case 12:
			featureRecord := FeatureLevelRecord{}
			featureRecord.frameVersion = record.frameVersion
			featureRecord.recordType = record.recordType
			featureRecord.version = record.version
			featureRecord.nameLength, _ = binary.ReadUvarint(valueBuf)
			featureRecord.name = string(valueBuf.Next(int(featureRecord.nameLength - 1)))
			_ = binary.Read(valueBuf, binary.BigEndian, &featureRecord.featureLevel)
			featureRecord.taggedFieldCount, _ = binary.ReadUvarint(valueBuf)
			record.FeatureLevelRecord = featureRecord
		}
		record.headerArrayCount, _ = binary.ReadUvarint(fileBuffer)
		clusterMetadata.records = append(clusterMetadata.records, &record)
	}

	return clusterMetadata, nil
}
