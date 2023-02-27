package extension

import jsoniter "github.com/json-iterator/go"

func Register() {
	jsoniter.ConfigCompatibleWithStandardLibrary.RegisterExtension(&u64AsStringCodec{})
	jsoniter.ConfigCompatibleWithStandardLibrary.RegisterExtension(&timeZoneCodec{})
	jsoniter.ConfigCompatibleWithStandardLibrary.RegisterExtension(&u64SliceAsStringsCodec{})
}
