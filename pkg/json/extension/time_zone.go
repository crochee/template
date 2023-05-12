package extension

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"

	jsoniter "github.com/json-iterator/go"
)

type timeZoneCodec struct {
	jsoniter.DummyExtension
}

func (extension *timeZoneCodec) UpdateStructDescriptor(structDescriptor *jsoniter.StructDescriptor) {
	for _, binding := range structDescriptor.Fields {
		fieldType := binding.Field.Type()
		if fieldType.Kind() == reflect.Struct && fieldType.String() == "time.Time" {
			timeFormat, ok := binding.Field.Tag().Lookup("time_format")
			if !ok {
				continue
			}
			if timeFormat == "" {
				timeFormat = time.RFC3339
			}
			timeLocation := binding.Field.Tag().Get("time_location")
			binding.Encoder = &funcEncoder{fun: func(ptr unsafe.Pointer, stream *jsoniter.Stream) {
				val := *((*time.Time)(ptr))
				if y := val.Year(); y < 0 || y >= 10000 {
					// RFC 3339 is clear that years are 4 digits exactly.
					// See golang.org/issue/4556#c15 for more discussion.
					stream.Error = errors.New("Time.MarshalJSON: year outside of range [0,9999]")
					return
				}
				b := make([]byte, 0, len(timeFormat)+2)
				b = append(b, '"')
				switch tf := strings.ToLower(timeFormat); tf {
				case "unix":
					strconv.AppendInt(b, val.Unix(), 10)
				case "unixnano":
					strconv.AppendInt(b, val.UnixNano(), 10)
				default:
					var l *time.Location
					switch timeLocation {
					case "Local":
						l = time.Local
					case "UTC", "":
						l = time.UTC
					default:
						loc, err := time.LoadLocation(timeLocation)
						if err != nil {
							stream.Error = err
							return
						}
						l = loc
					}
					b = val.In(l).AppendFormat(b, timeFormat)
				}
				b = append(b, '"')
				_, stream.Error = stream.Write(b)
			}}
			binding.Decoder = &funcDecoder{func(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
				val := iter.ReadString()
				if val == "null" || val == "" {
					return
				}
				switch tf := strings.ToLower(timeFormat); tf {
				case "unix", "unixnano":
					tv, err := strconv.ParseInt(val, 10, 64)
					if err != nil {
						iter.Error = err
						return
					}
					d := time.Duration(1)
					if tf == "unixnano" {
						d = time.Second
					}
					*((*time.Time)(ptr)) = time.Unix(tv/int64(d), tv%int64(d))
					return
				default:
					var l *time.Location
					switch timeLocation {
					case "Local":
						l = time.Local
					case "UTC", "":
						l = time.UTC
					default:
						loc, err := time.LoadLocation(timeLocation)
						if err != nil {
							iter.Error = err
							return
						}
						l = loc
					}
					t, err := time.ParseInLocation(timeFormat, val, l)
					if err != nil {
						iter.Error = err
						return
					}
					*((*time.Time)(ptr)) = t
				}
			}}
		}
	}
}
