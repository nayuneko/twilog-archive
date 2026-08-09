package text

import "html"

// UnescapeHTML unescapes HTML entities in a string iteratively (up to 3 passes)
// until no more entities remain to be unescaped.
func UnescapeHTML(s string) string {
	for i := 0; i < 3; i++ {
		next := html.UnescapeString(s)
		if next == s {
			break
		}
		s = next
	}
	return s
}
