/*
Copyright 2016 Stanislav Liberman
Copyright 2020 Evan Wies

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

package util

import (
	"strconv"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	assert.Equal(t, uint32(512), SemanticVersionCompose(0, 2, 0))
}

func TestMod3(t *testing.T) {
	tests := []uint64{
		0, 1, 2, 3, 100, 101, 102, 1 << 31, 1<<31 - 1, 1<<31 - 2,
	}
	for _, v := range tests {
		assert.Equal(t, FastMod3(v), int32(v%3))
	}
}

func TestMemcpy(t *testing.T) {
	dst := make([]byte, 102)
	src := make([]byte, 102)
	for i := range src {
		src[i] = byte(i)
	}
	Memcpy(uintptr(unsafe.Pointer(&dst[0])), uintptr(unsafe.Pointer(&src[0])), 102)
	assert.Equal(t, src, dst)
}

func BenchmarkMemcpy(b *testing.B) {
	dst := make([]byte, 4096)
	src := make([]byte, 4096)
	dstp := uintptr(unsafe.Pointer(&dst[0]))
	srcp := uintptr(unsafe.Pointer(&src[0]))
	b.ResetTimer()
	for _, k := range []int32{0, 2, 4, 8, 16, 64, 256, 1024, 4096} {
		b.Run(strconv.Itoa(int(k)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				Memcpy(dstp, srcp, k)
			}

		})
	}
}
