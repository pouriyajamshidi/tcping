// Package options handles the user input
package options

import (
	"testing"
	"time"

	"github.com/pouriyajamshidi/tcping/v3/internal/utils"
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

func TestSecondsToDuration(t *testing.T) {
	tests := []struct {
		name     string
		seconds  float64
		duration time.Duration
	}{
		{
			name:     "positive integer",
			seconds:  2,
			duration: 2 * time.Second,
		},
		{
			name:     "positive float",
			seconds:  1.5, // 1.5 = 3 / 2
			duration: time.Second * 3 / 2,
		},
		{
			name:     "negative integer",
			seconds:  -3,
			duration: -3 * time.Second,
		},
		{
			name:     "negative float",
			seconds:  -2.5, // -2.5 = -5 / 2
			duration: time.Second * -5 / 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.duration, utils.SecondsToDuration(tt.seconds))
		})
	}
}
