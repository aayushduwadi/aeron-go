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

package atomic

import (
	"bytes"
	"strconv"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

var buffer *Buffer

func TestMakeAtomicBuffer(t *testing.T) {
	var b *Buffer

	b = MakeBuffer()
	t.Logf("buf: %v", b)

	b = MakeBuffer(nil)
	t.Logf("buf: %v", b)

	arr := make([]byte, 64)

	b = MakeBuffer(arr)
	t.Logf("buf: %v", b)

	b = MakeBuffer(arr, 64)
	t.Logf("buf: %v", b)

	b = MakeBuffer(unsafe.Pointer(&arr[0]))
	t.Logf("buf: %v", b)

	b = MakeBuffer(unsafe.Pointer(&arr[0]), 64)
	t.Logf("buf: %v", b)

	b = MakeBuffer(uintptr(unsafe.Pointer(&arr[0])), 64)
	t.Logf("buf: %v", b)
}

func TestInit(t *testing.T) {
	buffer = MakeBuffer()
	assert.Equal(t, int32(0), buffer.Capacity())

	bytes := make([]byte, 32)
	bufLen := int32(len(bytes))

	buffer.Wrap(unsafe.Pointer(&bytes), bufLen)
	assert.Equal(t, bufLen, buffer.Capacity())
}

func TestGetAndAddInt64(t *testing.T) {
	buffer = NewBufferSlice(make([]byte, 32))
	buffer.Fill(0)

	assert.Equal(t, int64(0), buffer.GetAndAddInt64(0, 7))
	assert.Equal(t, int64(0), buffer.GetAndAddInt64(8, 7))
	assert.Equal(t, int64(0), buffer.GetAndAddInt64(16, 7))
	assert.Equal(t, int64(0), buffer.GetAndAddInt64(24, 7))

	assert.Equal(t, int64(7), buffer.GetAndAddInt64(0, 7))
	assert.Equal(t, int64(7), buffer.GetAndAddInt64(8, 7))
	assert.Equal(t, int64(7), buffer.GetAndAddInt64(16, 7))
	assert.Equal(t, int64(7), buffer.GetAndAddInt64(24, 7))

	assert.Equal(t, int64(14), buffer.GetAndAddInt64(0, 7))
	assert.Equal(t, int64(14), buffer.GetAndAddInt64(8, 7))
	assert.Equal(t, int64(14), buffer.GetAndAddInt64(16, 7))
	assert.Equal(t, int64(14), buffer.GetAndAddInt64(24, 7))
}

func TestPutInt64Ordered(t *testing.T) {
	buffer = NewBufferSlice(make([]byte, 32))
	buffer.Fill(0)

	buffer.PutInt64Ordered(1, 31415)
	assert.Equal(t, int64(31415), buffer.GetInt64Volatile(1))
	assert.NotEqual(t, int64(31415), buffer.GetInt64Volatile(2))
	assert.NotEqual(t, int64(31415), buffer.GetInt64Volatile(0))
}

func TestWrap(t *testing.T) {
	bytes := make([]byte, 32)
	ptr := unsafe.Pointer(&bytes[0])
	buffer = NewBufferSlice(bytes)
	buffer.Fill(0)
	t.Logf("buf: %v", buffer)

	buffer.PutInt64Ordered(1, 31415)
	assert.Equal(t, int64(31415), buffer.GetInt64Volatile(1))
	assert.NotEqual(t, int64(31415), buffer.GetInt64Volatile(2))
	assert.NotEqual(t, int64(31415), buffer.GetInt64Volatile(0))

	newPtr := unsafe.Pointer(uintptr(ptr) + uintptr(1))
	t.Logf("Old pointer: %v; new pointer: %v", ptr, newPtr)
	var newLen int32 = 31
	buffer.Wrap(newPtr, newLen)
	assert.Equal(t, newLen, buffer.Capacity())

	assert.NotEqual(t, int64(31415), buffer.GetInt64Volatile(1))
	assert.Equal(t, int64(31415), buffer.GetInt64Volatile(0))
}

func TestWriteBytes(t *testing.T) {
	buffer = NewBufferSlice(make([]byte, 32))
	buffer.PutInt8(1, 1)
	buffer.PutInt8(5, 5)

	var dest bytes.Buffer
	buffer.WriteBytes(&dest, 1, 5)
	assert.Equal(t, dest.Bytes(), []byte{1, 0, 0, 0, 5})
}

func BenchmarkWriteBytes(b *testing.B) {
	dst := bytes.NewBuffer(make([]byte, 4096))
	src := NewBufferSlice(make([]byte, 4096))
	b.ResetTimer()
	for _, k := range []int32{0, 2, 4, 8, 16, 64, 256, 1024, 4096} {
		b.Run(strconv.Itoa(int(k)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				dst.Reset()
				src.WriteBytes(dst, 0, k)
			}
		})
	}
}
