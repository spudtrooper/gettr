package util

import (
	"fmt"
	"strings"
)

func FormatNumber(n int) string {
	if n >= 1e12 {
		d := float64(n) / float64(1e12)
		return strings.Replace(fmt.Sprintf("%.1fT", d), ".0", "", 1)
	}
	if n >= 1e9 {
		d := float64(n) / float64(1e9)
		return strings.Replace(fmt.Sprintf("%.1fB", d), ".0", "", 1)
	}
	if n >= 1e6 {
		d := float64(n) / float64(1e6)
		return strings.Replace(fmt.Sprintf("%.1fM", d), ".0", "", 1)
	}
	if n >= 1e3 {
		d := float64(n) / float64(1e3)
		return strings.Replace(fmt.Sprintf("%.1fK", d), ".0", "", 1)
	}
	return fmt.Sprintf("%d", n)
}
