/**
 * @Author: zhangchao
 * @Description:
 * @Date: 2022/1/26 10:46 上午
 */

package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/second-state/WasmEdge-go/wasmedge"
)

type host struct {
	fetchResult []byte
}

// do the http fetch
func fetch(url string) []byte {
	resp, err := http.Get(string(url))
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	return body
}

// Host function for fetching
func (h *host) log(_ interface{}, mem *wasmedge.Memory, params []interface{}) ([]interface{}, wasmedge.Result) {
	// get url from memory
	pointer := params[0].(int32)
	size := params[1].(int32)
	data, _ := mem.GetData(uint(pointer), uint(size))

	log.Println(string(data))

	return nil, wasmedge.Result_Success
}

// Host function for fetching
func (h *host) hello(_ interface{}, mem *wasmedge.Memory, params []interface{}) ([]interface{}, wasmedge.Result) {
	// get url from memory
	pointer := params[0].(int32)
	size := params[1].(int32)
	data, _ := mem.GetData(uint(pointer), uint(size))

	res := "hello " + string(data)

	return nil, wasmedge.Result_Success
}

// Host function for writting memory
func (h *host) writeMem(_ interface{}, mem *wasmedge.Memory, params []interface{}) ([]interface{}, wasmedge.Result) {
	// write source code to memory
	pointer := params[0].(int32)
	mem.SetData(h.fetchResult, uint(pointer), uint(len(h.fetchResult)))

	return nil, wasmedge.Result_Success
}

func main() {
	fmt.Println("Go: Args:", os.Args)
	/// Expected Args[0]: program name (./externref)
	/// Expected Args[1]: wasm file (funcs.wasm)

	/// Set not to print debug info
	wasmedge.SetLogErrorLevel()

	conf := wasmedge.NewConfigure(wasmedge.WASI)
	vm := wasmedge.NewVMWithConfig(conf)
	obj := wasmedge.NewImportObject("env")

	h := host{}
	// Add host functions into the import object
	funcLogType := wasmedge.NewFunctionType(
		[]wasmedge.ValType{
			wasmedge.ValType_I32,
			wasmedge.ValType_I32,
		},
		[]wasmedge.ValType{
		})
	funcHelloType := wasmedge.NewFunctionType(
		[]wasmedge.ValType{
			wasmedge.ValType_I32,
			wasmedge.ValType_I32,
			wasmedge.ValType_I32,
			wasmedge.ValType_I32,
		},
		[]wasmedge.ValType{
			wasmedge.ValType_I32,
		})

	hostFetch := wasmedge.NewFunction(funcLogType, h.log, nil, 0)
	obj.AddFunction("log", hostFetch)
	hostFetch := wasmedge.NewFunction(funcLogType, h.log, nil, 0)
	obj.AddFunction("log", hostFetch)

	funcWriteType := wasmedge.NewFunctionType(
		[]wasmedge.ValType{
			wasmedge.ValType_I32,
		},
		[]wasmedge.ValType{})
	hostWrite := wasmedge.NewFunction(funcWriteType, h.writeMem, nil, 0)
	obj.AddFunction("write_mem", hostWrite)

	vm.RegisterImport(obj)

	vm.LoadWasmFile(os.Args[1])
	vm.Validate()
	vm.Instantiate()

	r, _ := vm.Execute("run")
	fmt.Printf("There are %d 'google' in source code of google.com\n", r[0])

	obj.Release()
	vm.Release()
	conf.Release()
}
