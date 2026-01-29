package slug

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

func makeSlug(title string) (slug string) {
	s := strings.ToLower(title)

	t := norm.NFD.String(s)

	var b strings.Builder
	for _, r := range t {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(r)
	}

	s = b.String()

	reg := regexp.MustCompile(`[^a-z0-9]+`)
	s = reg.ReplaceAllString(s, "-")

	s = strings.Trim(s, "-")

	return s
}

func SlugWithTime(title string) string {
	return fmt.Sprintf(
		"%s-%s",
		time.Now().Format("20060102-150405"),
		makeSlug(title),
	)
}
