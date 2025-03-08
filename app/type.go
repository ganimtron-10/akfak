package main

import (
	"bytes"
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

type RequestInterface interface {
	parse(buffer *bytes.Buffer)
}

type ResponseInterface interface {
	bytes(buffer *bytes.Buffer, includetagField bool)
}
