package extension

import (
	"testing"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
)

func Test_TimeUTCToShanghaiCodecEncode(t *testing.T) {
	Register()
	type TestString struct {
		T1 time.Time `json:"t_1"`
		T2 time.Time `json:"t_2,tzsh"`
	}

	tests := []struct {
		name    string
		input   TestString
		want    []byte
		wantErr bool
	}{
		{
			name: "ok",
			input: TestString{
				T1: time.Date(2020, 11, 13, 15, 16, 17, 18, time.UTC),
				T2: time.Date(2020, 11, 13, 15, 16, 17, 18, time.UTC),
			},
			want:    []byte(`{"t_1":"2020-11-13T15:16:17.000000018Z","t_2":"2020-11-13 15:16:17"}`),
			wantErr: false,
		},
		{
			name: "error",
			input: TestString{
				T1: time.Date(2020, 11, 13, 15, 16, 17, 18, time.UTC),
				T2: time.Date(-1, 11, 13, 15, 16, 17, 18, time.UTC),
			},
			want:    []byte(`{"t_1":"2020-11-13T15:16:17.000000018Z","t_2":"2020-11-13 15:16:17"}`),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_TimeUTCToShanghaiCodecDecode(t *testing.T) {
	Register()
	type TestString struct {
		T1 time.Time `json:"t_1"`
		T2 time.Time `json:"t_2,tzsh"`
	}

	tests := []struct {
		name    string
		input   []byte
		want    TestString
		wantErr bool
	}{
		{
			name:  "standard",
			input: []byte(`{"t_1":"2020-11-13T15:16:17.000000018Z","t_2":"2020-11-13 15:16:17"}`),
			want: TestString{
				T1: time.Date(2020, 11, 13, 15, 16, 17, 18, time.UTC),
				T2: time.Date(2020, 11, 13, 15, 16, 17, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name:  "error",
			input: []byte(`{"t_1":"2020-11-13T15:16:17.000000018Z","t_2":"2020-11-13 15:16:171"}`),
			want: TestString{
				T1: time.Date(2020, 11, 13, 15, 16, 17, 18, time.UTC),
				T2: time.Date(2020, 11, 13, 15, 16, 17, 0, time.UTC),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got TestString
			err := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(tt.input, &got)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
