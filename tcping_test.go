package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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

func TestNanoToMilliseconds(t *testing.T) {
	t.Parallel()
	tests := []struct {
		d    time.Duration
		want float32
	}{
		{d: time.Millisecond, want: 1},
		{d: 100*time.Millisecond + 123*time.Nanosecond, want: 100.000123},
		{d: time.Second, want: 1000},
		{d: time.Second + 100*time.Nanosecond, want: 1000.000123},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.d.String(), func(t *testing.T) {
			t.Parallel()
			got := nanoToMillisecond(tt.d.Nanoseconds())
			assert.Equal(t, tt.want, got)
		})
	}
}
