package validator

import "fmt"

func formatFloatDetail(name string, value float64, threshold float64, label string) string {
	return fmt.Sprintf("%s: %.2f (%s: %.2f)", name, value, label, threshold)
}

func formatIntDetail(name string, value int, threshold int, label string) string {
	return fmt.Sprintf("%s: %d (%s: %d)", name, value, label, threshold)
}
