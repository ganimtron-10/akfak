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

type ResponseHeader struct {
	messageSize   int32
	correlationId int32
}

// ApiVersions

type ApiVersionsRequest struct {
	RequestHeader
	clientSoftwareName    string
	clientSoftwareVersion string
}

type ApiKeys struct {
	apiKey     int16
	minVersion int16
	maxVersion int16
}

type ApiVersionsResponse struct {
	ResponseHeader
	errorCode    int16
	apiKeys      ApiKeys
	throttleTime int32
}

type RequestInterface interface {
	parse(buffer *bytes.Buffer)
}
