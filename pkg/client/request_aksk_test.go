package client

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_convertBody(t *testing.T) {
	c := &coPartner{
		ak:  "ak",
		sk:  "sk",
		req: OriginalRequest{},
	}
	type Inner struct {
		Content string
	}
	type Values struct {
		Key  string
		Data map[string]string
		A    Inner
		B    map[string]Inner
		C    map[string]map[string]Inner
		D    int
	}

	type args struct {
		body interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    map[string][]string
		wantErr bool
	}{
		{
			name:    "zero",
			args:    args{},
			want:    map[string][]string{},
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
					D: 6,
				},
			},
			want: map[string][]string{
				"A":    {"{\"Content\":\"aContent\"}"},
				"B":    {"{\"bKey1\":{\"Content\":\"b1Content\"},\"bKey2\":{\"Content\":\"b2Content\"}}"},
				"C":    {"{\"cKey1\":{\"ck11\":{\"Content\":\"ck11Content\"},\"ck12\":{\"Content\":\"ck12Content\"}},\"cKey2\":{\"ck21\":{\"Content\":\"ck21Content\"},\"ck22\":{\"Content\":\"ck22Content\"}}}"},
				"Data": {"{\"dataKey1\":\"dataValue1\",\"dataKey2\":\"dataValue2\"}"},
				"Key":  {"test"},
				"D":    {"6"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.convertBody(tt.args.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertBody() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_sign(t *testing.T) {
	h := make(http.Header)
	h.Set(HeaderKeySignedHeader, "0")
	c := &coPartner{
		ak:  "ak",
		sk:  "sk",
		req: OriginalRequest{},
	}
	type args struct {
		method    string
		resultMap map[string][]string
		header    http.Header
		body      interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "",
			args: args{
				method:    http.MethodPut,
				resultMap: map[string][]string{},
				header:    h,
				body: map[string]interface{}{
					"security_group_id": 439562762746635992,
					"direction":         "ingress",
					"ether_type":        "IPv4",
					"protocol":          "tcp",
					"port_range_min":    22,
					"port_range_max":    22,
					"description":       "默认安全组规则: 入方向, 放行22端口",
					"remote_group_id":   "",
					"remote_ip_prefix":  "0.0.0.0/0",
				},
			},
			want:    "ac0da47204211d0e3fa25271bbb0a9838c7557c065686c8a34419c5f82e8d8f4",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.sign(tt.args.method, tt.args.resultMap, tt.args.header, tt.args.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("sign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("sign() got = %v, want %v", got, tt.want)
			}
		})
	}
}
