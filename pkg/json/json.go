package json

import (
	"bytes"
	"encoding/json"
	"io"

	jsoniter "github.com/json-iterator/go"
)

type RawMessage json.RawMessage

func Marshal(input interface{}) ([]byte, error) {
	return jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(input)
}

func Unmarshal(input []byte, data interface{}) error {
	return jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(input, data)
}

func MarshalToString(input interface{}) (string, error) {
	return jsoniter.ConfigCompatibleWithStandardLibrary.MarshalToString(input)
}

func UnmarshalFromString(input string, data interface{}) error {
	return jsoniter.ConfigCompatibleWithStandardLibrary.UnmarshalFromString(input, data)
}

// UnmarshalNumber 当data没有指定具体数据结构时，json默认会将uint64数字转化为浮点数，这可能
// 导致精度丢失，使用该方法可以防止该问题出现
func UnmarshalNumber(input []byte, data interface{}) error {
	d := jsoniter.ConfigCompatibleWithStandardLibrary.NewDecoder(bytes.NewReader(input))
	d.UseNumber()
	return d.Decode(data)
}

// DecodeUseNumber 当data没有指定具体数据结构时，json默认会将uint64数字转化为浮点数，这可能
// 导致精度丢失，使用该方法可以防止该问题出现
func DecodeUseNumber(reader io.Reader, data interface{}) error {
	d := jsoniter.ConfigCompatibleWithStandardLibrary.NewDecoder(reader)
	d.UseNumber()
	return d.Decode(data)
}
