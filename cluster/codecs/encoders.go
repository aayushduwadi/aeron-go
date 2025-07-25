// Copyright (C) 2021-2022 Talos, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package codecs contains the archive protocol packet encoding and decoding
package codecs

import (
	"bytes"

	"github.com/lirm/aeron-go/aeron/atomic"
)

// Encoders for all cluster protocol packets
//
// Each of these functions creates a []byte suitable for sending over
// the wire by using the generated encoders created using
// simple-binary-encoding.
//
// All the packet specificationss are defined in the aeron-cluster protocol
// maintained at:
// http://github.com/real-logic/aeron/blob/master/aeron-cluster/src/main/resources/cluster/aeron-cluster-codecs.xml)
//
// The codecs are generated from that specification using Simple
// Binary Encoding (SBE) from https://github.com/real-logic/simple-binary-encoding

func ServiceAckRequestPacket(
	marshaller *SbeGoMarshaller,
	rangeChecking bool,
	logPosition int64,
	timestamp int64,
	ackID int64,
	relevantID int64,
	serviceID int32,
) ([]byte, error) {
	request := ServiceAck{
		LogPosition: logPosition,
		Timestamp:   timestamp,
		AckId:       ackID,
		RelevantId:  relevantID,
		ServiceId:   serviceID,
	}

	// Marshal it
	header := MessageHeader{
		BlockLength: request.SbeBlockLength(),
		TemplateId:  request.SbeTemplateId(),
		SchemaId:    request.SbeSchemaId(),
		Version:     request.SbeSchemaVersion(),
	}

	buffer := new(bytes.Buffer)
	if err := header.Encode(marshaller, buffer); err != nil {
		return nil, err
	}
	if err := request.Encode(marshaller, buffer, rangeChecking); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func SnapshotMarkerPacket(
	marshaller *SbeGoMarshaller,
	rangeChecking bool,
	typeId int64,
	logPosition int64,
	leadershipTermId int64,
	index int32,
	mark SnapshotMarkEnum,
	timeUnit ClusterTimeUnitEnum,
	appVersion int32,
) ([]byte, error) {
	request := SnapshotMarker{
		TypeId:           typeId,
		LogPosition:      logPosition,
		LeadershipTermId: leadershipTermId,
		Index:            index,
		Mark:             mark,
		TimeUnit:         timeUnit,
		AppVersion:       appVersion,
	}
	header := MessageHeader{
		BlockLength: request.SbeBlockLength(),
		TemplateId:  request.SbeTemplateId(),
		SchemaId:    request.SbeSchemaId(),
		Version:     request.SbeSchemaVersion(),
	}

	buffer := new(bytes.Buffer)
	if err := header.Encode(marshaller, buffer); err != nil {
		return nil, err
	}
	if err := request.Encode(marshaller, buffer, rangeChecking); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func ClientSessionPacket(
	marshaller *SbeGoMarshaller,
	rangeChecking bool,
	clusterSessionId int64,
	responseStreamId int32,
	responseChannel []byte,
	encodedPrincipal []byte,
) ([]byte, error) {
	request := ClientSession{
		ClusterSessionId: clusterSessionId,
		ResponseStreamId: responseStreamId,
		ResponseChannel:  responseChannel,
		EncodedPrincipal: encodedPrincipal,
	}
	header := MessageHeader{
		BlockLength: request.SbeBlockLength(),
		TemplateId:  request.SbeTemplateId(),
		SchemaId:    request.SbeSchemaId(),
		Version:     request.SbeSchemaVersion(),
	}

	buffer := new(bytes.Buffer)
	if err := header.Encode(marshaller, buffer); err != nil {
		return nil, err
	}
	if err := request.Encode(marshaller, buffer, rangeChecking); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func CloseSessionRequestPacket(
	marshaller *SbeGoMarshaller,
	rangeChecking bool,
	clusterSessionId int64,
) ([]byte, error) {
	request := CloseSession{
		ClusterSessionId: clusterSessionId,
	}

	// Marshal it
	header := MessageHeader{
		BlockLength: request.SbeBlockLength(),
		TemplateId:  request.SbeTemplateId(),
		SchemaId:    request.SbeSchemaId(),
		Version:     request.SbeSchemaVersion(),
	}

	buffer := new(bytes.Buffer)
	if err := header.Encode(marshaller, buffer); err != nil {
		return nil, err
	}
	if err := request.Encode(marshaller, buffer, rangeChecking); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func MakeClusterMessageBuffer(templateId, blockLength uint16) *atomic.Buffer {
	buf := atomic.NewBufferSlice(make([]byte, 8+blockLength))
	buf.PutUInt16(0, blockLength)
	buf.PutUInt16(2, templateId)
	buf.PutUInt16(4, 111) // schemaId
	buf.PutUInt16(6, 8)   // schemaVersion
	return buf
}
