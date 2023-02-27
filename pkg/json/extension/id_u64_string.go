package extension

import (
	"reflect"
	"strconv"
	"strings"
	"unsafe"

	"github.com/json-iterator/go"
)

type u64AsStringCodec struct {
	jsoniter.DummyExtension
}

func (extension *u64AsStringCodec) UpdateStructDescriptor(structDescriptor *jsoniter.StructDescriptor) {
	for _, binding := range structDescriptor.Fields {
		if binding.Field.Type().Kind() == reflect.Uint64 {
			tagParts := strings.Split(binding.Field.Tag().Get("json"), ",")
			if len(tagParts) <= 1 {
				continue
			}
			for _, tagPart := range tagParts[1:] {
				if tagPart == "string" {
					binding.Encoder = &funcEncoder{fun: func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
						val := *((*uint64)(ptr))
						var err error
						if val == 0 {
							_, err = stream.Write([]byte(nil))
						} else {
							_, err = stream.Write([]byte(strconv.FormatUint(val, 10)))
						}
						stream.Error = err
					}}
					binding.Decoder = &funcDecoder{func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
						if iter.WhatIsNext() != jsoniter.StringValue {
							*((*uint64)(ptr)) = iter.ReadUint64()
						}
					}}
					break
				}
			}
		}
	}
}
