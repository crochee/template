package extension

import (
	"reflect"
	"testing"

	"github.com/json-iterator/go"
)

func Test_Encode(t *testing.T) {
	Register()
	type TestString struct {
		ID     uint64 `json:"id,string"`
		Value  string `json:"value"`
		PID    uint64 `json:"pid"`
		PartID uint64 `json:",string"`
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
			want:    []byte(`{"id":"","value":"","pid":0,"PartID":""}`),
			wantErr: false,
		},
		{
			name: "num",
			input: TestString{
				ID: 787446465166,
			},
			want:    []byte(`{"id":"787446465166","value":"","pid":0,"PartID":""}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(tt.input)
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

func Test_Decode(t *testing.T) {
	Register()
	type TestString struct {
		ID     uint64 `json:"id,string"`
		Value  string `json:"value"`
		PID    uint64 `json:"pid"`
		PartID uint64 `json:",string"`
	}

	tests := []struct {
		name    string
		input   []byte
		want    TestString
		wantErr bool
	}{
		{
			name:  "standard",
			input: []byte(`{"id":"0","value":"","pid":0,"PartID":""}`),
			want: TestString{
				ID:     0,
				Value:  "",
				PID:    0,
				PartID: 0,
			},
			wantErr: false,
		},
		{
			name:  "none",
			input: []byte(`{"id":"","value":"","pid":0,"PartID":""}`),
			want: TestString{
				ID:     0,
				Value:  "",
				PID:    0,
				PartID: 0,
			},
			wantErr: false,
		},
		{
			name:  "num",
			input: []byte(`{"id":"787446465166","value":"","pid":0,"PartID":""}`),
			want: TestString{
				ID:     787446465166,
				Value:  "",
				PID:    0,
				PartID: 0,
			},
			wantErr: false,
		},
		{
			name:  "num",
			input: []byte(`{"id":"787446465166","value":"","pid":0,"PartID":"0"}`),
			want: TestString{
				ID:     787446465166,
				Value:  "",
				PID:    0,
				PartID: 0,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got TestString
			err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(tt.input, &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Unmarshal() got = %v, want %v", got, tt.want)
			}
		})
	}
}
