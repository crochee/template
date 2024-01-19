package code

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

func TestFroze(t *testing.T) {
	type args struct {
		code    string
		message string
	}
	tests := []struct {
		name string
		args args
		want ErrorCode
	}{
		{
			name: "nil",
			args: args{
				code:    "DSF.4000000002",
				message: "",
			},
			want: ErrInvalidParam,
		},
		{
			name: "nil",
			args: args{
				code:    "4000000002",
				message: "",
			},
			want: ErrInvalidParam,
		},
		{
			name: "nil",
			args: args{
				code:    "400-0000002",
				message: "",
			},
			want: ErrInvalidParam,
		},
		{
			name: "nil",
			args: args{
				code:    "0000002",
				message: "",
			},
			want: ErrInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Froze(tt.args.code, tt.args.message); !errors.Is(got, tt.want) {
				t.Errorf("Froze() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarshal(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		data, err := json.Marshal(ErrInternalServerError)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(data, []byte("{\"code\":\"5000000001\",\"message\":\"服务器内部错误\",\"result\":null}")) {
			t.Fatal(string(data))
		}
	})
	t.Run("old", func(t *testing.T) {
		data, err := json.Marshal(ErrCodeInternalServerError)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(data, []byte("{\"code\":\"5001000002\",\"message\":\"服务器内部错误\",\"result\":null}")) {
			t.Fatal(string(data))
		}
	})
}
