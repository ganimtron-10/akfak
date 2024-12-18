package main

import "bytes"

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

type RequestInterface interface {
	parse(buffer *bytes.Buffer)
}

type ResponseInterface interface {
	bytes(buffer *bytes.Buffer)
}
