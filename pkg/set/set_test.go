package set

import (
	"reflect"
	"testing"
)

func TestNewSet(t *testing.T) {
	type args struct {
		items []interface{}
	}
	tests := []struct {
		name string
		args args
		want *Set
	}{
		{
			name: "NewSet",
			args: args{
				items: []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			},
			want: &Set{map[interface{}]struct{}{1: {}, 2: {}, 3: {}, 4: {}, 5: {}, 6: {}, 7: {}, 8: {}, 9: {}, 10: {}}},
		},
		{
			name: "NewEmptySet",
			args: args{
				items: []interface{}{},
			},
			want: &Set{map[interface{}]struct{}{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSet(tt.args.items...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_IsContains(t *testing.T) {

	tests := []struct {
		Name  string
		Args  []string
		Value string
		Want  bool
	}{
		{
			Name:  "nullArgs",
			Args:  []string{},
			Value: "1",
			Want:  false,
		},
		{
			Name:  "yesArgs",
			Args:  []string{"1", "2", "3", "4"},
			Value: "3",
			Want:  true,
		},
		{
			Name:  "yesArgs",
			Args:  []string{"1", "2", "3", "4"},
			Value: "2",
			Want:  true,
		},
		{
			Name:  "noArgs",
			Args:  []string{"1", "2", "3", "4"},
			Value: "5",
			Want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			if !reflect.DeepEqual(IsContains(tt.Value, tt.Args), tt.Want) {
				t.Errorf("IsContains() = %v, want %v", IsContains(tt.Value, tt.Args), tt.Want)
			}
		})
	}
}

func Test_set_Add(t *testing.T) {
	type fields struct {
		m map[interface{}]struct{}
	}
	type args struct {
		items []interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Set
	}{
		{
			name: "AddEmptyOneItem",
			fields: fields{
				m: map[interface{}]struct{}{}},
			args: args{
				items: []interface{}{1},
			},
			want: &Set{map[interface{}]struct{}{1: {}}},
		},
		{
			name: "AddEmptyTwoItems",
			fields: fields{
				m: map[interface{}]struct{}{}},
			args: args{
				items: []interface{}{1, 2},
			},
			want: &Set{map[interface{}]struct{}{1: {}, 2: {}}},
		},
		{
			name: "AddExistOneItem",
			fields: fields{
				m: map[interface{}]struct{}{1: {}, 2: {}},
			},
			args: args{
				items: []interface{}{3},
			},
			want: &Set{map[interface{}]struct{}{1: {}, 2: {}, 3: {}}},
		},
		{
			name: "AddExistTwoItems",
			fields: fields{
				m: map[interface{}]struct{}{1: {}, 2: {}},
			},
			args: args{
				items: []interface{}{3, 4},
			},
			want: &Set{map[interface{}]struct{}{1: {}, 2: {}, 3: {}, 4: {}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Set{
				m: tt.fields.m,
			}
			s.Add(tt.args.items...)
			if !reflect.DeepEqual(s, tt.want) {
				t.Errorf("Set.Add() = %v, want %v", s, tt.want)
			}
		})
	}
}

func Test_set_Contains(t *testing.T) {
	type fields struct {
		m map[interface{}]struct{}
	}
	type args struct {
		item interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Contains",
			fields: fields{
				m: map[interface{}]struct{}{1: {}, 2: {}, 3: {}},
			},
			args: args{
				item: 1,
			},
			want: true,
		},
		{
			name: "NotContains",
			fields: fields{
				m: map[interface{}]struct{}{1: {}, 2: {}, 3: {}},
			},
			args: args{
				item: 4,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Set{
				m: tt.fields.m,
			}
			if got := s.Contains(tt.args.item); got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_set_Remove(t *testing.T) {
	type fields struct {
		m map[interface{}]struct{}
	}
	type args struct {
		item interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Set
	}{
		{
			name: "Remove",
			fields: fields{
				m: map[interface{}]struct{}{1: {}, 2: {}, 3: {}},
			},
			args: args{
				item: 1,
			},
			want: &Set{map[interface{}]struct{}{2: {}, 3: {}}},
		},
		{
			name: "NotRemove",
			fields: fields{
				m: map[interface{}]struct{}{1: {}, 2: {}, 3: {}},
			},
			args: args{
				item: 4,
			},
			want: &Set{map[interface{}]struct{}{1: {}, 2: {}, 3: {}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Set{
				m: tt.fields.m,
			}
			s.Remove(tt.args.item)
			if !reflect.DeepEqual(s, tt.want) {
				t.Errorf("Set.Remove() = %v, want %v", s, tt.want)
			}
		})
	}
}

func Test_set_Size(t *testing.T) {
	type fields struct {
		m map[interface{}]struct{}
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name:   "SizeEmpty",
			fields: fields{m: map[interface{}]struct{}{}},
			want:   0,
		},
		{
			name: "SizeNotEmpty",
			fields: fields{
				m: map[interface{}]struct{}{1: {}, 2: {}, 3: {}},
			},
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Set{
				m: tt.fields.m,
			}
			if got := s.Size(); got != tt.want {
				t.Errorf("Size() = %v, want %v", got, tt.want)
			}
		})
	}
}
