package extension

import (
	"reflect"
	"testing"

	jsoniter "github.com/json-iterator/go"
)

func Test_ignoreForeignKeyCodec_UpdateStructDescriptor(t *testing.T) {
	j := jsoniter.Config{
		EscapeHTML:             true,
		SortMapKeys:            true,
		ValidateJsonRawMessage: true,
	}.Froze()

	RegisterWith(j, &IgnoreForeignKeyCodec{})
	type FixedIP struct {
		ID    uint64 `json:"id,string"`
		Value string `json:"value"`
	}
	type TestString struct {
		ID       uint64    `json:"id,string"`
		Value    string    `json:"value"`
		PID      uint64    `json:"pid"`
		PartID   uint64    `json:",string"`
		FixedIPs []FixedIP `gorm:"foreignKey:NetCardID;references:ID" json:"fixed_ips"`
	}

	tests := []struct {
		name    string
		input   TestString
		want    []byte
		wantErr bool
	}{
		{
			name: "none",
			input: TestString{
				ID:       0,
				Value:    "",
				PID:      0,
				PartID:   0,
				FixedIPs: []FixedIP{},
			},
			want:    []byte(`{"id":"0","value":"","pid":0,"PartID":"0","fixed_ips":null}`),
			wantErr: false,
		},
		{
			name: "num",
			input: TestString{
				ID:     787446465166,
				Value:  "",
				PID:    0,
				PartID: 0,
				FixedIPs: []FixedIP{
					{
						ID:    2,
						Value: "f",
					},
				},
			},
			want:    []byte(`{"id":"787446465166","value":"","pid":0,"PartID":"0","fixed_ips":null}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := j.Marshal(tt.input)
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
