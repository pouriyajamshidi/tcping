package main

import "testing"

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
