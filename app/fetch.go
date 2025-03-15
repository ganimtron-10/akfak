package main

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/google/uuid"
)

// Fetch

type FetchPartition struct {
	Partition          int32
	CurrentLeaderEpoch int32
	FetchOffset        int64
	LastFetchedEpoch   int32
	LogStartOffset     int64
	PartitionMaxBytes  int32
}

type FetchTopic struct {
	TopicID    uuid.UUID
	Partitions []*FetchPartition
}

type ForgottenTopicData struct {
	TopicID    uuid.UUID
	Partitions []int32
}

type FetchRequest struct {
	RequestHeader
	MaxWaitMs           int32
	MinBytes            int32
	MaxBytes            int32
	IsolationLevel      int8
	SessionID           int32
	SessionEpoch        int32
	Topics              []*FetchTopic
	ForgottenTopicsData []*ForgottenTopicData
	RackID              string
}

type AbortedTransaction struct {
	ProducerID  int64
	FirstOffset int64
}

type FetchResponsePartition struct {
	PartitionIndex       int32
	ErrorCode            int16
	HighWatermark        int64
	LastStableOffset     int64
	LogStartOffset       int64
	AbortedTransactions  []*AbortedTransaction
	PreferredReadReplica int32
	Records              []byte
}

type FetchResponseTopic struct {
	TopicID    uuid.UUID
	Partitions []*FetchResponsePartition
}

type FetchResponse struct {
	ThrottleTimeMs int32
	ErrorCode      int16
	SessionID      int32
	Responses      []*FetchResponseTopic
}

func (request *FetchRequest) parse(buffer *bytes.Buffer) {

	binary.Read(buffer, binary.BigEndian, &request.MaxWaitMs)
	binary.Read(buffer, binary.BigEndian, &request.MinBytes)
	binary.Read(buffer, binary.BigEndian, &request.MaxBytes)
	binary.Read(buffer, binary.BigEndian, &request.IsolationLevel)
	binary.Read(buffer, binary.BigEndian, &request.SessionID)
	binary.Read(buffer, binary.BigEndian, &request.SessionEpoch)

	topicsLength, _ := binary.ReadUvarint(buffer)

	request.Topics = make([]*FetchTopic, topicsLength-1)
	for i := 0; i < int(topicsLength-1); i++ {
		topic := &FetchTopic{}
		topicBytes := make([]byte, 16)
		buffer.Read(topicBytes)
		topic.TopicID, _ = uuid.FromBytes(topicBytes)
		partitionsLength, _ := binary.ReadUvarint(buffer)

		topic.Partitions = make([]*FetchPartition, partitionsLength-1)

		for j := 0; j < int(partitionsLength-1); j++ {
			partition := &FetchPartition{}
			binary.Read(buffer, binary.BigEndian, &partition.Partition)
			binary.Read(buffer, binary.BigEndian, &partition.CurrentLeaderEpoch)
			binary.Read(buffer, binary.BigEndian, &partition.FetchOffset)
			binary.Read(buffer, binary.BigEndian, &partition.LastFetchedEpoch)
			binary.Read(buffer, binary.BigEndian, &partition.LogStartOffset)
			binary.Read(buffer, binary.BigEndian, &partition.PartitionMaxBytes)
			topic.Partitions[j] = partition
		}
		request.Topics[i] = topic
	}

	forgottenTopicsLength, _ := binary.ReadUvarint(buffer)
	if forgottenTopicsLength > 0 {
		request.ForgottenTopicsData = make([]*ForgottenTopicData, forgottenTopicsLength-1)
		for i := 0; i < int(forgottenTopicsLength-1); i++ {
			forgottenTopic := &ForgottenTopicData{}
			topicBytes := make([]byte, 16)
			buffer.Read(topicBytes)
			forgottenTopic.TopicID, _ = uuid.FromBytes(topicBytes)
			forgottenPartitionsLength, _ := binary.ReadUvarint(buffer)

			forgottenTopic.Partitions = make([]int32, forgottenPartitionsLength)
			for j := 0; j < int(forgottenPartitionsLength); j++ {
				binary.Read(buffer, binary.BigEndian, &forgottenTopic.Partitions[j])
			}
			request.ForgottenTopicsData[i] = forgottenTopic
		}
	}

	rackIDLength, _ := binary.ReadUvarint(buffer)
	if rackIDLength > 0 {
		rackIDBytes := make([]byte, rackIDLength-1)
		buffer.Read(rackIDBytes)

		request.RackID = string(rackIDBytes)
	}

	ignoreTagField(buffer)
	fmt.Printf("%+v\n", request)
}

func (response *FetchResponse) bytes(buffer *bytes.Buffer) {

	binary.Write(buffer, binary.BigEndian, response.ThrottleTimeMs)
	binary.Write(buffer, binary.BigEndian, response.ErrorCode)
	binary.Write(buffer, binary.BigEndian, response.SessionID)

	binary.Write(buffer, binary.BigEndian, binary.AppendUvarint([]byte{}, uint64(len(response.Responses)+1)))
	for _, topicResponse := range response.Responses {
		topicBytes, _ := topicResponse.TopicID.MarshalBinary()
		buffer.Write(topicBytes)

		binary.Write(buffer, binary.BigEndian, binary.AppendUvarint([]byte{}, uint64(len(topicResponse.Partitions)+1)))
		for _, partitionResponse := range topicResponse.Partitions {
			binary.Write(buffer, binary.BigEndian, partitionResponse.PartitionIndex)
			binary.Write(buffer, binary.BigEndian, partitionResponse.ErrorCode)
			binary.Write(buffer, binary.BigEndian, partitionResponse.HighWatermark)
			binary.Write(buffer, binary.BigEndian, partitionResponse.LastStableOffset)
			binary.Write(buffer, binary.BigEndian, partitionResponse.LogStartOffset)

			binary.Write(buffer, binary.BigEndian, binary.AppendUvarint([]byte{}, uint64(len(partitionResponse.AbortedTransactions)+1)))
			for _, abortedTransaction := range partitionResponse.AbortedTransactions {
				binary.Write(buffer, binary.BigEndian, abortedTransaction.ProducerID)

				binary.Write(buffer, binary.BigEndian, abortedTransaction.FirstOffset)
				addTagField(buffer)
			}

			binary.Write(buffer, binary.BigEndian, partitionResponse.PreferredReadReplica)

			binary.Write(buffer, binary.BigEndian, binary.AppendUvarint([]byte{}, uint64(len(partitionResponse.Records)+1)))
			buffer.Write(partitionResponse.Records)
			addTagField(buffer)
		}
		addTagField(buffer)
	}
	addTagField(buffer)

}

func (request *FetchRequest) generateResponse(commonResponse *Response) {
	commonResponse.correlationId = request.correlationId

	fetchResponse := FetchResponse{}
	fetchResponse.ThrottleTimeMs = 0
	fetchResponse.ErrorCode = 0
	fetchResponse.SessionID = request.SessionID

	for _, topic := range request.Topics {
		topicResponse := &FetchResponseTopic{
			TopicID: topic.TopicID,
			Partitions: []*FetchResponsePartition{
				{
					PartitionIndex: 0,
					ErrorCode:      0,
				},
			},
		}
		fetchResponse.Responses = append(fetchResponse.Responses, topicResponse)
	}

	fetchResponse.bytes(&commonResponse.BytesData)
}
