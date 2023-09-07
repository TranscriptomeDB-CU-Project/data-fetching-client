package utils

import (
	"regexp"
)

var re = regexp.MustCompile(`([^"]+)\.sdrf\.txt`)

func ExtractSDRFFileName(body string) []string {
	var result []string

	matches := re.FindAllStringSubmatch(body, -1)

	for _, match := range matches {
		result = append(result, match[0])
	}

	return result
}
