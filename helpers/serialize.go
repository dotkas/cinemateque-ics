package helpers

import (
	"bytes"
	"fmt"
	"strings"
)

func MustSerialize(v string) string {
	b := bytes.NewBufferString("")

	if strings.ContainsAny(v, ";:\\\",") {
		v = strings.Replace(v, ",", "\\,", -1)
		v = strings.Replace(v, "\"", "\\\"", -1)
		v = strings.Replace(v, "\\", "\\\\", -1)
	}

	if len(v) > 75 {
		if _, err := fmt.Fprintln(b, v[:75]); err != nil {
			panic(err)
		}
		v = v[75:]

		if _, err := fmt.Fprint(b, " "); err != nil {
			panic(err)

		}
	}

	for len(v) > 74 {
		if _, err := fmt.Fprintln(b, v[:74]); err != nil {
			panic(err)

		}
		v = v[74:]

		if _, err := fmt.Fprint(b, " "); err != nil {
			panic(err)
		}
	}

	if _, err := fmt.Fprintln(b, v); err != nil {
		panic(err)

	}

	return b.String()
}
