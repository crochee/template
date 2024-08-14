package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type structCustomValidation struct {
	Order string `binding:"order"`
}

func TestOrderWithDBSort(t *testing.T) {
	engine, err := New()
	assert.Nil(t, err)
	err = RegisterValidation(engine, "order", OrderWithDBSort)
	assert.Nil(t, err)

	type args struct {
		value string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "nil",
			args: args{
				value: "",
			},
			wantErr: true,
		},
		{
			name: "one_ok",
			args: args{
				value: "update",
			},
			wantErr: false,
		},
		{
			name: "one_single_char",
			args: args{
				value: "u",
			},
			wantErr: true,
		},
		{
			name: "one_single_invalid_char",
			args: args{
				value: "%",
			},
			wantErr: true,
		},
		{
			name: "one_two_char_contain_",
			args: args{
				value: "u_",
			},
			wantErr: true,
		},
		{
			name: "one_many_char",
			args: args{
				value: "ucfd 98he'	s'[",
			},
			wantErr: true,
		},
		{
			name: "mult_ok",
			args: args{
				value: "io asc,created_at,updated_at ASC,name desc,deleted DESC",
			},
			wantErr: false,
		},
		{
			name: "mult_failed",
			args: args{
				value: "upDate DESC,created",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = engine.ValidateStruct(structCustomValidation{Order: tt.args.value})
			if tt.wantErr {
				assert.NotNil(t, err, "err:%v", err)
				return
			}
			assert.Nil(t, err, "err:%v", err)
		})
	}
}

func TestComponentName(t *testing.T) {
	engine, err := New()
	assert.Nil(t, err)
	err = RegisterValidation(engine, "component_name", ComponentName)
	assert.Nil(t, err)

	type Component struct {
		Name string `binding:"required,component_name"`
	}
	type args struct {
		value string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "",
			args: args{},
			want: true,
		},
		{
			name: "legal",
			args: args{
				value: "89sd_合法",
			},
			want: false,
		},
		{
			name: "limit_length",
			args: args{
				value: "89sd_合法89sd_合法89sd_合法89sd_合法89sd_合法89sd_合法89sd_合法89sd_合法89sd_合法",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = engine.ValidateStruct(Component{Name: tt.args.value})
			if tt.want {
				assert.NotNil(t, err, "err:%v", err)
				return
			}
			assert.Nil(t, err, "err:%v", err)
		})
	}
}

type CommaListValidateTestReq struct {
	RequestTypes string `form:"request_types" binding:"comma_list=cpu mem"`
}

func TestCommaListValidate(t *testing.T) {
	engine, err := New()
	assert.Nil(t, err)
	err = RegisterValidation(engine, "comma_list", CommaListValidate)
	assert.Nil(t, err)

	type args struct {
		value string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "normal",
			args: args{
				value: "cpu mem",
			},
			want: true,
		},
		{
			name: "contains not exist element",
			args: args{
				value: "cpu mem test",
			},
			want: false,
		},
		{
			name: "contains duplicated element",
			args: args{
				value: "cpu cpu cpu",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = engine.ValidateStruct(CommaListValidateTestReq{RequestTypes: tt.args.value})
			if (err != nil) != tt.want {
				assert.NotNil(t, err, "err:%v", err)
				return
			}
		})
	}
}
