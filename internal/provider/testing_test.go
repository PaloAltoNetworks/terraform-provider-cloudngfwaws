package provider

import (
	"strings"
)

func sliceToString(v []string) string {
	var buf strings.Builder

	buf.WriteString("[")
	for i, x := range v {
		if i > 0 {
			buf.WriteString(", ")
		}

		buf.WriteString("\"")
		buf.WriteString(x)
		buf.WriteString("\"")
	}
	buf.WriteString("]")

	return buf.String()
}
