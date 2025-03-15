package main

import (
	"bytes"
	"encoding/binary"
	"fmt"

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

func (request *DescribePartitionsRequest) parse(buffer *bytes.Buffer) {
	request.names = getStringArray(buffer)
	binary.Read(buffer, binary.BigEndian, &request.responsePartitionLimit)
	ignoreTagField(buffer)
	fmt.Printf("%+v\n", request)
}

func (response *DescribePartitionsResponse) bytes(buffer *bytes.Buffer, request *DescribePartitionsRequest) {

	binary.Write(buffer, binary.BigEndian, response.throttleTime)

	topicsToSend := []*Topic{}

	// for _, topic := range response.topics {
	// 	if slices.Contains(request.names, topic.name) {
	// 		topicsToSend = append(topicsToSend, topic)
	// 	}
	// }

	for _, requestTopicName := range request.names {
		foundTopic := false
		for _, topic := range response.topics {
			if topic.name == requestTopicName {
				topicsToSend = append(topicsToSend, topic)
				foundTopic = true
			}
		}

		if !foundTopic {
			topicsToSend = append(topicsToSend, &Topic{errorCode: 3, name: requestTopicName})
		}
	}

	binary.Write(buffer, binary.BigEndian, int8(len(topicsToSend)+1))
	for _, topic := range topicsToSend {
		binary.Write(buffer, binary.BigEndian, topic.errorCode)
		writeCompactString(buffer, topic.name)

		binary.Write(buffer, binary.BigEndian, topic.topicId[:])
		binary.Write(buffer, binary.BigEndian, topic.isInternal)

		// fmt.Printf("\nPartition Record in bytes: %+v", topic.partitions)

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

				// fmt.Printf("\nTopic details: %+v", topic)
				// fmt.Printf("\nPartition Record adding in response: %+v", record.PartitionRecord)
			case 12:
				// featureRecord
				// Skipping for now
			}
		}
	}

	return nil
}
