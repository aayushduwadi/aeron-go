/*
Copyright 2016 Stanislav Liberman

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package logbuffer

const DataFrameHeader_FrameLengthFieldOffset int32 = 0
const DataFrameHeader_VersionFieldOffset int32 = 4
const DataFrameHeader_FlagsFieldOffset int32 = 5
const DataFrameHeader_TypeFieldOffset int32 = 6
const DataFrameHeader_TermOffsetFieldOffset int32 = 8
const DataFrameHeader_SessionIDFieldOffset int32 = 12
const DataFrameHeader_StreamIDFieldOffset int32 = 16
const DataFrameHeader_TermIDFieldOffset int32 = 20
const DataFrameHeader_ReservedValueFieldOffset int32 = 24
const DataFrameHeader_DataOffset int32 = 32

const DataFrameHeader_Length int32 = 32

const DataFrameHeader_TypePad uint16 = 0x00
const DataFrameHeader_TypeData uint16 = 0x01
const DataFrameHeader_TypeNAK uint16 = 0x02
const DataFrameHeader_TypeSM uint16 = 0x03
const DataFrameHeader_TypeErr uint16 = 0x04
const DataFrameHeader_TypeSetup uint16 = 0x05
const DataFrameHeader_TypeExt uint16 = 0xFFFF

const DataFrameHeader_CurrentVersion int8 = 0x0

var DataFrameHeader = struct {
	FrameLengthFieldOffset   int32
	VersionFieldOffset       int32
	FlagsFieldOffset         int32
	TypeFieldOffset          int32
	TermOffsetFieldOffset    int32
	SessionIDFieldOffset     int32
	StreamIDFieldOffset      int32
	TermIDFieldOffset        int32
	ReservedValueFieldOffset int32
	DataOffset               int32

	Length int32

	TypePad   uint16
	TypeData  uint16
	TypeNAK   uint16
	TypeSM    uint16
	TypeErr   uint16
	TypeSetup uint16
	TypeExt   uint16

	CurrentVersion int8
}{
	DataFrameHeader_FrameLengthFieldOffset,
	DataFrameHeader_VersionFieldOffset,
	DataFrameHeader_FlagsFieldOffset,
	DataFrameHeader_TypeFieldOffset,
	DataFrameHeader_TermOffsetFieldOffset,
	DataFrameHeader_SessionIDFieldOffset,
	DataFrameHeader_StreamIDFieldOffset,
	DataFrameHeader_TermIDFieldOffset,
	DataFrameHeader_ReservedValueFieldOffset,
	DataFrameHeader_DataOffset,

	DataFrameHeader_Length,

	DataFrameHeader_TypePad,
	DataFrameHeader_TypeData,
	DataFrameHeader_TypeNAK,
	DataFrameHeader_TypeSM,
	DataFrameHeader_TypeErr,
	DataFrameHeader_TypeSetup,
	DataFrameHeader_TypeExt,

	DataFrameHeader_CurrentVersion,
}
