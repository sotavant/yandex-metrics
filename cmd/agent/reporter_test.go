package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_getURL(t *testing.T) {
	type args struct {
		mType string
		name  string
		value string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: `url`,
			args: args{
				mType: `a`,
				name:  `b`,
				value: `c`,
			},
			want: `/update/a/b/c`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, getURL(tt.args.mType, tt.args.name, tt.args.value), "getURL(%v, %v, %v)", tt.args.mType, tt.args.name, tt.args.value)
		})
	}
}
