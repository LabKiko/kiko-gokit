package prom

import (
	"strings"
)

func BuildMetric(names ...string) string {
	var b strings.Builder
	for i := 0; i < len(names); i++ {
		if names[i] != "" {
			if b.Len() > 0 {
				b.WriteString("_")
			}
			b.WriteString(names[i])
		}
	}

	return b.String()
}
