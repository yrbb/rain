package utils

import (
	"unsafe"
)

func String(i []byte) string {
	if len(i) == 0 {
		return ""
	}

	return *(*string)(unsafe.Pointer(&i))
}

func Slice(i string) []byte {
	if i == "" {
		return nil
	}

	return *(*[]byte)(unsafe.Pointer(&i))
}
