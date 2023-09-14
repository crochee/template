package extension

import jsoniter "github.com/json-iterator/go"

func Register() {
	RegisterWith(jsoniter.ConfigCompatibleWithStandardLibrary,
		&U64AsStringCodec{},
		&TimeZoneCodec{},
		&U64SliceAsStringsCodec{},
	)
}

func RegisterWith(api jsoniter.API, extensions ...jsoniter.Extension) {
	for _, extension := range extensions {
		api.RegisterExtension(extension)
	}
}
