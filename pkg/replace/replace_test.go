package replace

import (
	"regexp"
	"testing"
)

func TestMessageReplacer_Replace(t *testing.T) {
	type fields struct {
		replace *regexp.Regexp
	}
	type args struct {
		input string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "nil",
			fields: fields{
				replace: regexp.MustCompile(`(?U)"(password|admin_pass)":"(.*)",?`),
			},
			args: args{
				input: `"key":"iiii","password":"","name":"value"`,
			},
			want: `"key":"iiii","password":"******","name":"value"`,
		},
		{
			name: "password",
			fields: fields{
				replace: regexp.MustCompile(`(?U)"(password|admin_pass)":"(.*)",?`),
			},
			args: args{
				input: `"key":"iiii","password":"uiuiui","name":"value"`,
			},
			want: `"key":"iiii","password":"******","name":"value"`,
		},
		{
			name: "admin_pass",
			fields: fields{
				replace: regexp.MustCompile(`(?U)"(password|admin_pass)":"(.*)",?`),
			},
			args: args{
				input: `"key":"iiii","admin_pass":"uiuiui","name":"value"`,
			},
			want: `"key":"iiii","admin_pass":"******","name":"value"`,
		},
		{
			name: "admin_pass not ,",
			fields: fields{
				replace: regexp.MustCompile(`(?U)"(password|admin_pass)":"(.*)",?`),
			},
			args: args{
				input: `"key":"iiii","admin_pass":"uiuiui"`,
			},
			want: `"key":"iiii","admin_pass":"******"`,
		},
		{
			name: "admin_pass not , more",
			fields: fields{
				replace: regexp.MustCompile(`(?U)"(password|admin_pass)":"(.*)",?`),
			},
			args: args{
				input: `{"key":"iiii","admin_pass":"uiuiui"}`,
			},
			want: `{"key":"iiii","admin_pass":"******"}`,
		},
		{
			name: "admin_pass nil",
			fields: fields{
				replace: regexp.MustCompile(`(?U)"(password|admin_pass)":"(.*)",?`),
			},
			args: args{
				input: `{"key":"iiii","admin_pass":""}`,
			},
			want: `{"key":"iiii","admin_pass":"******"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := messageReplacer{
				match: tt.fields.replace,
			}
			if got := m.Replace(tt.args.input); got != tt.want {
				t.Errorf("Replace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNopReplacer_Replace(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "op",
			args: args{
				"lll",
			},
			want: "lll",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			no := NopReplacer{}
			if got := no.Replace(tt.args.input); got != tt.want {
				t.Errorf("Replace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPasswordReplacer_Replace(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "null",
			args: args{
				input: "",
			},
			want: "******",
		},
		{
			name: "big",
			args: args{
				input: "sgdvhjajkna knbdvvskdv",
			},
			want: "******",
		},
		{
			name: "short",
			args: args{
				input: "sg",
			},
			want: "******",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PasswordReplacer{}
			if got := p.Replace(tt.args.input); got != tt.want {
				t.Errorf("Replace() = %v, want %v", got, tt.want)
			}
		})
	}
}
