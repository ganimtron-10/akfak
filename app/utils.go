package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// func calculateMessageSize(message []byte) int32 {
// 	return int32(len(message))
// }

// func getInt32FromBytes(buffer []byte, offset int32) int32 {
// 	return int32(binary.BigEndian.Uint32(buffer[offset : offset+4+1]))
// }

// func getInt16FromBytes(buffer []byte, offset int32) int16 {
// 	return int16(binary.BigEndian.Uint16(buffer[offset : offset+2+1]))
// }

// func getBytesfromInt32(value int32) []byte {
// 	byteArray := []byte{}
// 	byteArray = binary.BigEndian.AppendUint32(byteArray, uint32(value))
// 	return byteArray
// }

// func getBytesfromInt16(value int16) []byte {
// 	byteArray := []byte{}
// 	byteArray = binary.BigEndian.AppendUint16(byteArray, uint16(value))
// 	return byteArray
// }

func readCompactString(buffer *bytes.Buffer) string {
	strLength, err := binary.ReadUvarint(buffer)
	if err != nil {
		fmt.Println("Error reading COMPACT_STRING length: ", err.Error())
	}
	str := string(buffer.Next(int(strLength - 1)))
	return str
}

func readNullableString(buffer *bytes.Buffer) string {
	var strLength int16
	err := binary.Read(buffer, binary.BigEndian, &strLength)
	if err != nil {
		fmt.Println("Error reading NULLABLE_STRING length: ", err.Error())
	}
	str := string(buffer.Next(int(strLength)))
	return str
}

func getStringArray(buffer *bytes.Buffer) []string {
	stringArray := []string{}
	arrayLength, err := binary.ReadVarint(buffer)
	if err != nil {
		fmt.Println("Error reading ARRAY length: ", err.Error())
	}

	for i := 0; i < int(arrayLength); i++ {
		stringArray = append(stringArray, readCompactString(buffer))
		ignoreTagField(buffer)
	}

	return stringArray
}

func ignoreTagField(buffer *bytes.Buffer) {
	// not an ideal implementation, will figure this out later
	_, err := binary.ReadUvarint(buffer)
	if err != nil {
		fmt.Println("Error reading TAGGED_FIELD length: ", err.Error())
	}
}

func addTagField(buffer *bytes.Buffer) {
	// not an ideal implementation, will figure this out later
	_, err := buffer.Write([]byte{0})
	if err != nil {
		fmt.Println("Error writing TAGGED_FIELD: ", err.Error())
	}
}

func getApiVersionsErrorCode(apiVersion int16) int16 {
	switch apiVersion {
	case 0, 1, 2, 3, 4:
		return 0
	default:
		return 35
	}
}
