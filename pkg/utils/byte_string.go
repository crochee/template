package utils

import (
	"reflect"
	"unsafe"
)

// String cost-free conversion of byte stream to character
func String(b []byte) (s string) {
	pBytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	pString := (*reflect.StringHeader)(unsafe.Pointer(&s))
	pString.Data = pBytes.Data
	pString.Len = pBytes.Len
	return
}

// Bytes cost-free conversion of characters to byte streams
func Bytes(s string) (b []byte) {
	pBytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	pString := (*reflect.StringHeader)(unsafe.Pointer(&s))
	pBytes.Data = pString.Data
	pBytes.Len = pString.Len
	pBytes.Cap = pString.Len
	return
}
