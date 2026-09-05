package cli

import "testing"

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name string
		v1   string
		v2   string
		want int
	}{
		{
			name: "equal versions",
			v1:   "1.2.3",
			v2:   "1.2.3",
			want: 0,
		},
		{
			name: "v1 major less than v2",
			v1:   "1.2.3",
			v2:   "2.0.0",
			want: -1,
		},
		{
			name: "v1 major greater than v2",
			v1:   "3.0.0",
			v2:   "2.9.9",
			want: 1,
		},
		{
			name: "v1 minor less than v2",
			v1:   "1.2.3",
			v2:   "1.3.0",
			want: -1,
		},
		{
			name: "v1 minor greater than v2",
			v1:   "1.4.0",
			v2:   "1.3.9",
			want: 1,
		},
		{
			name: "v1 patch less than v2",
			v1:   "1.2.3",
			v2:   "1.2.4",
			want: -1,
		},
		{
			name: "v1 patch greater than v2",
			v1:   "1.2.5",
			v2:   "1.2.4",
			want: 1,
		},
		{
			name: "v1 shorter but equal in common parts",
			v1:   "1.2",
			v2:   "1.2.0",
			want: -1,
		},
		{
			name: "v1 longer but equal in common parts",
			v1:   "1.2.0",
			v2:   "1.2",
			want: 1,
		},
		{
			name: "v1 shorter and less in common parts",
			v1:   "1.2",
			v2:   "1.3.0",
			want: -1,
		},
		{
			name: "v1 longer and greater in common parts",
			v1:   "2.0.0",
			v2:   "1.9",
			want: 1,
		},
		{
			name: "single-digit versions equal",
			v1:   "1",
			v2:   "1",
			want: 0,
		},
		{
			name: "non-numeric segment treated as zero",
			v1:   "1.x.0",
			v2:   "1.0.0",
			want: 0,
		},
		{
			name: "both empty strings",
			v1:   "",
			v2:   "",
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareVersions(tt.v1, tt.v2)
			if got != tt.want {
				t.Errorf("compareVersions(%q, %q) = %d, want %d", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}
