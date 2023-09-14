package extension

import (
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
)

type U64SliceAsStringsCodec struct {
	jsoniter.DummyExtension
}

func (extension *U64SliceAsStringsCodec) UpdateStructDescriptor(structDescriptor *jsoniter.StructDescriptor) {
	for _, binding := range structDescriptor.Fields {
		if binding.Field.Type().Kind() == reflect.Slice {
			tagParts := strings.Split(binding.Field.Tag().Get("json"), ",")
			if len(tagParts) <= 1 {
				continue
			}
			for _, tagPart := range tagParts[1:] {
				if tagPart == "strings" {
					binding.Encoder = &funcEncoder{fun: func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
						nums := *((*[]uint64)(ptr))
						strs := make([]string, 0, len(nums))
						for _, num := range nums {
							strs = append(strs, strconv.FormatUint(num, 10))
						}
						stream.WriteVal(strs)
					}}
					binding.Decoder = &funcDecoder{
						fun: func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
							if iter.WhatIsNext() == jsoniter.ArrayValue {
								arr := []uint64{}
								iter.ReadArrayCB(func(iterator *jsoniter.Iterator) bool {
									value := iterator.ReadString()
									res, _ := strconv.ParseUint(value, 10, 64)
									arr = append(arr, res)
									return true
								})
								*((*[]uint64)(ptr)) = arr
							}
						},
					}
					break
				}
			}
		}
	}
}
