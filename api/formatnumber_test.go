package api

import (
	"fmt"
	"testing"
)

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  string
	}{
		{
			input: 1,
			want:  "1",
		},
		{
			input: 12,
			want:  "12",
		},
		{
			input: 123,
			want:  "123",
		},
		{
			input: 1234,
			want:  "1.2K",
		},
		{
			input: 12345,
			want:  "12.3K",
		},
		{
			input: 123456,
			want:  "123.5K",
		},
		{
			input: 1234567,
			want:  "1.2M",
		},
		{
			input: 12345678,
			want:  "12.3M",
		},
		{
			input: 123456789,
			want:  "123.5M",
		},
		{
			input: 1234567891,
			want:  "1.2B",
		},
		{
			input: 12345678912,
			want:  "12.3B",
		},
		{
			input: 123456789123,
			want:  "123.5B",
		},
		{
			input: 1234567891234,
			want:  "1.2T",
		},
		{
			input: 12345678912345,
			want:  "12.3T",
		},
		{
			input: 123456789123456,
			want:  "123.5T",
		},
		{
			input: 1234567891234567,
			want:  "1234.6T",
		},
	}
	for _, test := range tests {
		name := test.name
		if name == "" {
			name = fmt.Sprintf("%d", test.input)
		}
		t.Run(name, func(t *testing.T) {
			formatted := formatNumber(test.input)
			if got, want := formatted, test.want; got != want {
				t.Errorf("got(%q) != want(%q)", got, want)
			}
		})
	}
}
