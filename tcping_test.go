package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalcTime(t *testing.T) {
	type args struct {
		time uint
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "1 second",
			args: args{1},
			want: "1 second",
		},
		{
			name: "59 seconds",
			args: args{59},
			want: "59 seconds",
		},
		{
			name: "1 minute",
			args: args{1 * 60},
			want: "1 minute",
		},
		{
			name: "59 minutes 0 seconds",
			args: args{59 * 60},
			want: "59 minutes 0 seconds",
		},
		{
			name: "1 minute 5 seconds",
			args: args{1*60 + 5},
			want: "1 minute 5 seconds",
		},
		{
			name: "59 minutes 5 seconds",
			args: args{59*60 + 5},
			want: "59 minutes 5 seconds",
		},
		{
			name: "1 hour",
			args: args{1 * 60 * 60},
			want: "1 hour",
		},
		{
			name: "1 hour 10 minutes 5 seconds",
			args: args{1*60*60 + 10*60 + 5},
			want: "1 hour 10 minutes 5 seconds",
		},
		{
			name: "59 hours 0 minutes 0 seconds",
			args: args{59 * 60 * 60},
			want: "59 hours 0 minutes 0 seconds",
		},
		{
			name: "59 hours 10 minutes 5 seconds",
			args: args{59*60*60 + 10*60 + 5},
			want: "59 hours 10 minutes 5 seconds",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calcTime(tt.args.time); got != tt.want {
				t.Errorf("calcTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPermuteArgs(t *testing.T) {
	type args struct {
		args []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			"host/ip before option",
			args{args: []string{"127.0.0.1", "8080", "-r", "3"}},
			[]string{"-r", "3", "127.0.0.1", "8080"},
		},
		{
			"host/ip after option",
			args{args: []string{"-r", "3", "127.0.0.1", "8080"}},
			[]string{"-r", "3", "127.0.0.1", "8080"},
		},
		{
			"check for updates",
			args{args: []string{"-u"}},
			[]string{"-u"},
		},
		/**
		 * cases in which the value of the option does not exist are not listed.
		 * they call directly usage() and exit with code 1.
		 */
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permuteArgs(tt.args.args)
			assert.Equal(t, tt.want, tt.args.args)
		})
	}
}

func TestResolveHostname (t *testing.T) {
    validAddr := "127.0.0.1"
    s := &stats{hostname: validAddr}
    addr, err := resolveHostname(s)
    assert.Nil(t, err)
    assert.Equal(t, validAddr, addr.String())

    invalidAddr := "8.0.0.8.5"
    s = &stats{hostname: invalidAddr}
    addr, err = resolveHostname(s)
    assert.NotNil(t, err)
    assert.Equal(t, "invalid IP", addr.String())
}
