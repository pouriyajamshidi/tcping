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
		name    string
		args    args
		want    []string
		wantErr error
	}{
		{
			"host/ip before option",
			args{args: []string{"127.0.0.1", "8080", "-r", "3", "-u"}},
			[]string{"-r", "3", "-u", "127.0.0.1", "8080"},
			nil,
		},
		{
			"host/ip after option",
			args{args: []string{"-r", "3", "-u", "127.0.0.1", "8080"}},
			[]string{"-r", "3", "-u", "127.0.0.1", "8080"},
			nil,
		},
		{
			"args not enough",
			args{args: []string{"-r", "3", "-u", "127.0.0.1"}},
			[]string{"-r", "3", "-u", "127.0.0.1"},
			errArgsNotEnough,
		},
		{
			"args not enough. it means that the 8080 was interpreted as the value of -r",
			args{args: []string{"-u", "127.0.0.1", "-r", "8080"}},
			[]string{"-u", "127.0.0.1", "-r", "8080"},
			errArgsNotEnough,
		},
		{
			"flag has no value. the next flag has come",
			args{args: []string{"-r", "-u", "127.0.0.1", "8080"}},
			[]string{"-r", "-u", "127.0.0.1", "8080"},
			errArgsFlagHasNoValue,
		},
		{
			"flag has no value. out of index",
			args{args: []string{"-u", "127.0.0.1", "8080", "-r"}},
			[]string{"-u", "127.0.0.1", "8080", "-r"},
			errArgsFlagHasNoValue,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := permuteArgs(tt.args.args)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			}
			assert.Equal(t, tt.want, tt.args.args)
		})
	}
}
