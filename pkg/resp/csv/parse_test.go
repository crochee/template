package csv

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type Response struct {
	Code   int
	Msg    string
	Result interface{}
}

type Lists struct {
	List []*Content `csv:""`
}

type Content struct {
	Name   string    `csv:""`
	Age    int       `csv:"age,string"`
	Scores float64   `csv:"scores"`
	Create time.Time `csv:"-"`
	Inner  `csv:""`
	Other  *Other `csv:""`
	Value  interface{}
	point  int
}

type Inner struct {
	Color string `csv:""`
	Tool  string `csv:""`
}

type Other struct {
	Color string `csv:"Ocolor"`
	Tool  string `csv:"Otool"`
}

type Operate struct {
	Listener []*Other `csv:",fmt"`
}

type MapOp struct {
	List map[string]interface{} `csv:",dynamic_tile"`
}

func Test_parse_Parse(t *testing.T) {
	type fields struct {
		tagName string
	}
	type args struct {
		obj interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*mapIndexValue
		wantErr bool
	}{
		{
			name: "OK",
			fields: fields{
				tagName: "csv",
			},
			args: args{
				obj: []*Content{
					{
						Name:   "lihua",
						Age:    26,
						Create: time.Now(),
						Inner: Inner{
							Color: "89",
							Tool:  "pen",
						},
						Other: &Other{
							Color: "op",
							Tool:  "hss",
						},
					},
					{
						Name: "zhangsan",
						Age:  20,
					},
				},
			},
			want: []*mapIndexValue{
				{
					data: map[string]interface{}{
						"age":    "26",
						"Name":   "lihua",
						"scores": 0.0,
						"Color":  "89",
						"Tool":   "pen",
						"Ocolor": "op",
						"Otool":  "hss",
					},
				},
				{
					data: map[string]interface{}{
						"age":    "20",
						"Name":   "zhangsan",
						"scores": 0.0,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "slice",
			fields: fields{
				tagName: "csv",
			},
			args: args{
				obj: &Operate{
					Listener: []*Other{
						{
							Color: "p0",
							Tool:  "t0",
						},
						{
							Color: "p1",
							Tool:  "t1",
						},
						{
							Color: "p2",
							Tool:  "t2",
						},
					},
				},
			},
			want: []*mapIndexValue{
				{
					data: map[string]interface{}{
						"Listener(Ocolor)": "p0,p1,p2",
						"Listener(Otool)":  "t0,t1,t2",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "map dynamic_tile",
			fields: fields{
				tagName: "csv",
			},
			args: args{
				obj: MapOp{List: map[string]interface{}{
					"cpu": 9.0,
					"mem": 80,
				}},
			},
			want: []*mapIndexValue{
				{
					data: map[string]interface{}{
						"cpu": 9.0,
						"mem": 80,
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &parse{
				tagName: tt.fields.tagName,
			}
			got, err := p.parse(tt.args.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func TestFormat(t *testing.T) {
	type args struct {
		name  string
		input []*mapIndexValue
	}
	tests := []struct {
		name string
		args args
		want []*mapIndexValue
	}{
		{
			name: "ok",
			args: args{
				name: "test",
				input: []*mapIndexValue{
					{
						data: map[string]interface{}{
							"port": "8080",
							"host": "127.0.0.1",
						},
					},
					{
						data: map[string]interface{}{
							"port": "8081",
							"host": "127.0.0.2",
						},
					},
					{
						data: map[string]interface{}{
							"port": "8082",
							"host": "127.0.0.3",
						},
					},
				},
			},
			want: []*mapIndexValue{
				{
					data: map[string]interface{}{
						"test(port)": "8080,8081,8082",
						"test(host)": "127.0.0.1,127.0.0.2,127.0.0.3",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := format(tt.args.name, tt.args.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("format() = %v, want %v", got, tt.want)
			}
		})
	}
}
