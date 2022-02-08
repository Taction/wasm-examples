/**
 * @Author: zhangchao
 * @Description:
 * @Date: 2022/1/24 5:07 下午
 */
package main

import (
	"reflect"
	"unsafe"
)

func main() {
	data := callhosthello("from wasm")
	logstr(stringBytePtr(data), len(data))
}

//export logstr
func logstr(data *byte, len int)

//export hello
func hello(data *byte, size int, returnValueData **byte, returnValueSize *int) int32

func callhosthello(data string) string {
	var raw *byte
	var size int
	success := hello(stringBytePtr(data), len(data), &raw, &size)
	if success == 0 {
		// failed
		return ""
	}
	return RawBytePtrToString(raw, size)
}

func stringBytePtr(msg string) *byte {
	if len(msg) == 0 {
		return nil
	}
	bt := *(*[]byte)(unsafe.Pointer(&msg))
	return &bt[0]
}

func RawBytePtrToString(raw *byte, size int) string {
	return *(*string)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(raw)),
		Len:  uintptr(size),
		Cap:  uintptr(size),
	}))
}

func RawBytePtrToByteSlice(raw *byte, size int) []byte {
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(raw)),
		Len:  uintptr(size),
		Cap:  uintptr(size),
	}))
}