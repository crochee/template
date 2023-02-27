package code

import (
	"errors"
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
