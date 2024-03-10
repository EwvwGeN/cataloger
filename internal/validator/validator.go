package validator

import "regexp"

func ValideteByRegex(str, patern string) bool {
	regex := regexp.MustCompile(patern)
	return regex.MatchString(str)
}