package messagecontrol

import "testing"

func TestParseLimit(t *testing.T) {
	tests := []struct {
		input, def, max, expected int
	}{
		{0, 10, 100, 10},
		{-5, 10, 100, 10},
		{5, 10, 100, 5},
		{150, 10, 100, 100},
	}

	for _, tt := range tests {
		got := parseLimit(tt.input, tt.def, tt.max)
		if got != tt.expected {
			t.Errorf("parseLimit(%d, %d, %d) = %d; want %d", tt.input, tt.def, tt.max, got, tt.expected)
		}
	}
}
