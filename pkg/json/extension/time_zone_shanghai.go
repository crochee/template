package extension

import (
	"errors"
	"reflect"
	"strings"
	"time"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
)

var TimeFormatLayout = "2006-01-02 15:04:05"

type timeUTCToShanghaiCodec struct {
	jsoniter.DummyExtension
}

func (extension *timeUTCToShanghaiCodec) UpdateStructDescriptor(structDescriptor *jsoniter.StructDescriptor) {
	for _, binding := range structDescriptor.Fields {
		fieldType := binding.Field.Type()
		if fieldType.Kind() == reflect.Struct && fieldType.String() == "time.Time" {
			tagParts := strings.Split(binding.Field.Tag().Get("json"), ",")
			if len(tagParts) <= 1 {
				continue
			}
			for _, tagPart := range tagParts[1:] {
				if tagPart == "tzsh" {
					binding.Encoder = &funcEncoder{fun: func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
						val := *((*time.Time)(ptr))
						if y := val.Year(); y < 0 || y >= 10000 {
							// RFC 3339 is clear that years are 4 digits exactly.
							// See golang.org/issue/4556#c15 for more discussion.
							stream.Error = errors.New("Time.MarshalJSON: year outside of range [0,9999]")
							return
						}

						b := make([]byte, 0, len(TimeFormatLayout)+2)
						b = append(b, '"')
						b = val.AppendFormat(b, TimeFormatLayout)
						b = append(b, '"')
						_, _ = stream.Write(b)
					}}
					binding.Decoder = &funcDecoder{func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
						val := iter.ReadString()
						if val == "null" {
							return
						}
						t, err := time.Parse(TimeFormatLayout, val)
						if err != nil {
							iter.Error = err
							return
						}
						*((*time.Time)(ptr)) = t
					}}
					break
				}
			}
		}
	}
}
