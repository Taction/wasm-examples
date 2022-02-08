/**
 * @Author: zhangchao
 * @Description:
 * @Date: 2022/1/24 5:07 下午
 */
package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/wasmerio/wasmer-go/wasmer"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func main() {

	pwd, _ := os.Getwd()
	path := filepath.Join(pwd, "./wasm.wasm")

	wasm, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		check(err)
	}
	// Create an Engine
	engine := wasmer.NewEngine()

	// Create a Store
	store := wasmer.NewStore(engine)

	// Once we have our binary `wasm` we can compile that into a `*Module`
	// which represents compiled JIT code.
	module, err := wasmer.NewModule(store, wasm)
	check(err)

	var importObject *wasmer.ImportObject

	wasiEnv, err := wasmer.NewWasiStateBuilder("").Finalize()
	if err != nil || wasiEnv == nil {
		check(fmt.Errorf("new wasi err %w", err))
	}

	importObject, err = wasiEnv.GenerateImportObject(store, module)
	check(err)

	var instance *wasmer.Instance

	importObject.Register(
		"env",
		map[string]wasmer.IntoExtern{
			"hello": wasmer.NewFunction(
				store,
				wasmer.NewFunctionType(wasmer.NewValueTypes(wasmer.I32, wasmer.I32, wasmer.I32, wasmer.I32), wasmer.NewValueTypes(wasmer.I32)),
				func(args []wasmer.Value) ([]wasmer.Value, error) {
					in, bufferType, returnBufferData, returnBufferSize := args[0].I32(), args[1].I32(), args[2].I32(), args[3].I32()
					m, err := instance.Exports.GetMemory("memory")
					if err != nil {
						return nil, err
					}
					mem := m.Data()
					msg := string(mem[in:in+bufferType])
					res := "hello " + msg

					blen := int32(len(res))

					var addrIndex int32
					malloc, err := instance.Exports.GetRawFunction("malloc")
					if err == nil {
						addr, err := malloc.Call(blen)
						if err != nil {
							return []wasmer.Value{wasmer.NewI32(0)}, nil
						}
						addrIndex , _ = addr.(int32)
					} else {
						return []wasmer.Value{wasmer.NewI32(0)}, errors.New("malloc error")
					}

					copy(mem[addrIndex:], res[:])
					binary.LittleEndian.PutUint32(mem[returnBufferSize:], uint32(blen))
					binary.LittleEndian.PutUint32(mem[returnBufferData:], uint32(addrIndex))
					return []wasmer.Value{wasmer.NewI32(1)}, nil
				},
			),
			"logstr": wasmer.NewFunction(
				store,
				wasmer.NewFunctionType(wasmer.NewValueTypes(wasmer.I32, wasmer.I32), wasmer.NewValueTypes()),
				func(args []wasmer.Value) ([]wasmer.Value, error) {
					in, bufferSize := args[0].I32(), args[1].I32()
					m, err := instance.Exports.GetMemory("memory")
					if err != nil {
						return nil, err
					}
					mem := m.Data()
					msg := string(mem[in:in+bufferSize])
					log.Println(msg)
					return []wasmer.Value{}, nil
				},
			),
		},
	)

	// Next up we instantiate a module which is where we link in all our
	// imports. We've got one import so we pass that in here.
	instance, err = wasmer.NewInstance(module, importObject)
	check(err)

	start, err := instance.Exports.GetFunction("_start")
	check(err)
	start()
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}


