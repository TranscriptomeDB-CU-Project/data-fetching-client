package utils

import (
	"regexp"
	"strings"
)

var re = regexp.MustCompile(`("[^"path" : "]+)\.sdrf\.txt`)

func ExtractSDRFFileName(body string) []string {
	var result []string
	matches := re.FindAllStringSubmatch(body, -1)

	for _, match := range matches {
		result = append(result, strings.TrimLeft(match[0], "\""))
	}

	return result
}
