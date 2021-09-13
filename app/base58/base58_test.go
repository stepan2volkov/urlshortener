package base58

import (
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		name string
		want int
	}{
		{"0", 0},
		{"100", 100},
		{"76003", 76003},
		{"2000000000", 2000000000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoded, err := Decode(tt.want)
			if err != nil {
				t.Errorf("error got while decoding: %v", err)
			}

			got, err := Encode(decoded)
			if err != nil {
				t.Errorf("error got while encoding: %v", err)
			}
			if got != tt.want {
				t.Errorf("expected %d, want %d", got, tt.want)
			}
		})
	}
}
