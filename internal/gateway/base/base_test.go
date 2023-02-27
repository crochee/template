package base

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_convertBody(t *testing.T) {
	type Inner struct {
		Content string
	}
	type Values struct {
		Key  string
		Data map[string]string
		A    Inner
		B    map[string]Inner
		C    map[string]map[string]Inner
	}

	type args struct {
		body interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			name:    "zero",
			args:    args{},
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name: "more-success",
			args: args{
				body: Values{
					Key:  "test",
					Data: map[string]string{"dataKey1": "dataValue1", "dataKey2": "dataValue2"},
					A: Inner{
						Content: "aContent",
					},
					B: map[string]Inner{
						"bKey1": {
							Content: "b1Content",
						},
						"bKey2": {
							Content: "b2Content",
						},
					},
					C: map[string]map[string]Inner{
						"cKey1": {
							"ck11": {
								Content: "ck11Content",
							},
							"ck12": {
								Content: "ck12Content",
							},
						},
						"cKey2": {
							"ck21": {
								Content: "ck21Content",
							},
							"ck22": {
								Content: "ck22Content",
							},
						},
					},
				},
			},
			want: map[string]string{
				"A":    "{\"Content\":\"aContent\"}",
				"B":    "{\"bKey1\":{\"Content\":\"b1Content\"},\"bKey2\":{\"Content\":\"b2Content\"}}",
				"C":    "{\"cKey1\":{\"ck11\":{\"Content\":\"ck11Content\"},\"ck12\":{\"Content\":\"ck12Content\"}},\"cKey2\":{\"ck21\":{\"Content\":\"ck21Content\"},\"ck22\":{\"Content\":\"ck22Content\"}}}",
				"Data": "{\"dataKey1\":\"dataValue1\",\"dataKey2\":\"dataValue2\"}",
				"Key":  "\"test\""},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := (coPartner{}).convertBody(tt.args.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertBody() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
