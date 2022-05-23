package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
			name: "59.0 minutes.seconds",
			args: args{59 * 60},
			want: "59.0 minutes.seconds",
		},
		{
			name: "1.5 minute.seconds",
			args: args{1*60 + 5},
			want: "1.5 minute.seconds",
		},
		{
			name: "59.5 minutes.seconds",
			args: args{59*60 + 5},
			want: "59.5 minutes.seconds",
		},
		{
			name: "1 hour",
			args: args{1 * 60 * 60},
			want: "1 hour",
		},
		{
			name: "1.10.5 hour.minutes.seconds",
			args: args{1*60*60 + 10*60 + 5},
			want: "1.10.5 hour.minutes.seconds",
		},
		{
			name: "59.0.0 hours.minutes.seconds",
			args: args{59 * 60 * 60},
			want: "59.0.0 hours.minutes.seconds",
		},
		{
			name: "59.10.5 hours.minutes.seconds",
			args: args{59*60*60 + 10*60 + 5},
			want: "59.10.5 hours.minutes.seconds",
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
			args{args: []string{"127.0.0.1", "-p", "8080"}},
			[]string{"-p", "8080", "127.0.0.1"},
		},
		{
			"host/ip after option",
			args{args: []string{"-p", "8080", "127.0.0.1"}},
			[]string{"-p", "8080", "127.0.0.1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			permuteArgs(tt.args.args)
			assert.Equal(t, tt.want, tt.args.args)
		})
	}
}
