package utils

import (
	"errors"
	"fmt"
	"strings"
)

const pathParameterPrefix = ":"

// Converts a pattern like /users or /user/:id to file names
func PathPatternToFileName(pattern string) (string, error) {
	if len(pattern) == 0 {
		return "", errors.New("invalid pattern")
	}
	fname := &strings.Builder{}
	parts := strings.Split(pattern, "/")
	// make sure we can handle /users:id and users/:id
	if parts[0] == "" {
		parts = parts[1:]
	}

	for i, p := range parts {
		if strings.HasPrefix(p, pathParameterPrefix) {
			// convert :id to $id
			fname.WriteString(fmt.Sprintf("$%s", p[1:]))
		} else {
			fname.WriteString(p)
		}

		if i < len(parts)-1 {
			fname.WriteString(".")
		}
	}
	fname.WriteString(".ts")

	return fname.String(), nil
}

// Converts file names like users.$id.ts to users/:id etc
func FileNameToHttpPath(name string) string {
	return "/" + strings.ReplaceAll(
		strings.ReplaceAll(strings.TrimSuffix(name, ".ts"), "$", ":"),
		".", "/",
	)
}
