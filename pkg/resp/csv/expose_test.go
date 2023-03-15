package csv

import (
	"reflect"
	"testing"
)

func Test_expose_GetStruct(t *testing.T) {
	type args struct {
		data       interface{}
		fieldNames []string
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "OK",
			args: args{
				data: Response{
					Code: 500,
					Msg:  "err",
					Result: &Lists{List: []*Content{
						{
							Name: "zhangsan",
							Age:  20,
						},
						{
							Name: "lisi",
							Age:  22,
						},
					}},
				},
				fieldNames: []string{"Result", "List"},
			},
			want: []*Content{
				{
					Name: "zhangsan",
					Age:  20,
				},
				{
					Name: "lisi",
					Age:  22,
				},
			},
			wantErr: false,
		},
		{
			name: "no struct",
			args: args{
				data: Response{
					Code: 500,
					Msg:  "err",
					Result: Lists{List: []*Content{
						{
							Name: "zhangsan",
							Age:  20,
						},
						{
							Name: "lisi",
							Age:  22,
						},
					}},
				},
				fieldNames: []string{"Result", "List", "Age"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "nil",
			args: args{
				data: Response{
					Code:   500,
					Msg:    "err",
					Result: Lists{List: nil},
				},
				fieldNames: []string{"Result", "List", "Age"},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := expose{}
			got, err := e.GetStruct(tt.args.data, tt.args.fieldNames...)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStruct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetStruct() got = %v, want %v", got, tt.want)
			}
		})
	}
}
