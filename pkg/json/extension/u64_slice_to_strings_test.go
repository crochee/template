package extension

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/json-iterator/go"
)

func Test_U64s_Encode(t *testing.T) {
	type TestString struct {
		Foo []uint64 `json:"foo,strings"`
		Bar []uint64 `json:"bar"`
	}

	tests := []struct {
		name    string
		input   TestString
		want    []byte
		wantErr bool
	}{
		{
			name:    "none",
			input:   TestString{},
			want:    []byte(`{"foo":[],"bar":null}`),
			wantErr: false,
		},
		{
			name: "normal",
			input: TestString{
				Foo: []uint64{330120267575640153, 330120472828100697, 331857912358027388},
				Bar: []uint64{330120267575640153, 330120472828100697, 331857912358027388},
			},
			want:    []byte(`{"foo":["330120267575640153","330120472828100697","331857912358027388"],"bar":[330120267575640153,330120472828100697,331857912358027388]}`),
			wantErr: false,
		},
		{
			name: "oneItem",
			input: TestString{
				Foo: []uint64{330120267575640153},
				Bar: []uint64{330120267575640153, 330120472828100697, 331857912358027388},
			},
			want:    []byte(`{"foo":["330120267575640153"],"bar":[330120267575640153,330120472828100697,331857912358027388]}`),
			wantErr: false,
		},
	}

	json := jsoniter.ConfigCompatibleWithStandardLibrary
	json.RegisterExtension(&u64SliceAsStringsCodec{})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Marshal() got = %s, want %s", got, tt.want)
			}
		})
	}
}

func Test_U64s_Decode(t *testing.T) {
	type TestString struct {
		Foo []uint64 `json:"foo,strings"`
		Bar []uint64 `json:"bar"`
	}

	tests := []struct {
		name    string
		input   []byte
		want    TestString
		wantErr bool
	}{
		{
			name:    "none",
			input:   []byte(`{"foo":[],"bar":null}`),
			want:    TestString{},
			wantErr: false,
		},
		{
			name: "normal",
			want: TestString{
				Foo: []uint64{330120267575640153, 330120472828100697, 331857912358027388},
				Bar: []uint64{330120267575640153, 330120472828100697, 331857912358027388},
			},
			input:   []byte(`{"foo":["330120267575640153","330120472828100697","331857912358027388"],"bar":[330120267575640153,330120472828100697,331857912358027388]}`),
			wantErr: false,
		},
		{
			name: "oneItem",
			want: TestString{
				Foo: []uint64{330120267575640153},
				Bar: []uint64{330120267575640153, 330120472828100697, 331857912358027388},
			},
			input:   []byte(`{"foo":["330120267575640153"],"bar":[330120267575640153,330120472828100697,331857912358027388]}`),
			wantErr: false,
		},
	}

	json := jsoniter.ConfigCompatibleWithStandardLibrary
	json.RegisterExtension(&u64SliceAsStringsCodec{})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got TestString
			err := json.Unmarshal(tt.input, &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println("got:", got)
			fmt.Println("want", tt.want)
		})
	}
}
