package contesta

import "strings"

func Article(word string) string {
	if strings.HasPrefix(word, "a") ||
		strings.HasPrefix(word, "e") ||
		strings.HasPrefix(word, "i") ||
		strings.HasPrefix(word, "o") ||
		strings.HasPrefix(word, "u") {
		return "an"
	}
	return "a"
}
