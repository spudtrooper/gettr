package util

import "fmt"

func FormatNumber(n int) string {
	if n > 1e12 {
		d := float64(n) / float64(1e12)
		return fmt.Sprintf("%.1fT", d)
	}
	if n > 1e9 {
		d := float64(n) / float64(1e9)
		return fmt.Sprintf("%.1fB", d)
	}
	if n > 1e6 {
		d := float64(n) / float64(1e6)
		return fmt.Sprintf("%.1fM", d)
	}
	if n > 1e3 {
		d := float64(n) / float64(1e3)
		return fmt.Sprintf("%.1fK", d)
	}
	return fmt.Sprintf("%d", n)
}
