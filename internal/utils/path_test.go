package utils_test

import (
	"testing"

	"github.com/protomoks/pmok/internal/utils"
)

func TestPathPatternToFileName(t *testing.T) {

	cases := []struct {
		pattern   string
		want      string
		wantError string
	}{
		{
			pattern: "/users/:id",
			want:    "users.$id.ts",
		},
		{
			pattern: "users/:id",
			want:    "users.$id.ts",
		},
		{
			pattern:   "",
			wantError: "invalid pattern",
		},
		{
			pattern: "/oh/:id/my/god/:what",
			want:    "oh.$id.my.god.$what.ts",
		},
		{
			pattern: "/some/png.png",
			want:    "some.png.png.ts",
		},
	}

	getErr := func(err error) string {
		if err != nil {
			return err.Error()
		}
		return ""
	}

	for _, c := range cases {
		t.Run(c.pattern, func(t *testing.T) {
			res, err := utils.PathPatternToFileName(c.pattern)
			if getErr(err) != c.wantError {
				t.Fatalf("expected errors do not match. expected %s, but got %s", c.wantError, getErr(err))
			}
			if res != c.want {
				t.Fatalf("expected filename result to be %s, but got %s", c.want, res)
			}

		})
	}
}
