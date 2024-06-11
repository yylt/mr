package util

import (
	"bytes"
	"sync"
	"unsafe"
)

var (
	bufpool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
)

func GetBuf() *bytes.Buffer {
	buf := bufpool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

func PutBuf(b *bytes.Buffer) {
	bufpool.Put(b)
}

func Str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	b := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&b))
}

func Bytes2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func SetAppend(src1, src2 []string) []string {
	if len(src1) == 0 {
		return src2
	}
	if len(src2) == 0 {
		return src1
	}
	var (
		map1 = make(map[string]struct{}, len(src1))
	)
	for _, v := range src1 {
		map1[v] = struct{}{}

	}
	for _, v2 := range src2 {
		_, ok := map1[v2]
		if !ok {
			src1 = append(src1, v2)
		}
	}
	return src1
}
