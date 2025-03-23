package util

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func ToTitleCase(s string) string {
	caser := cases.Title(language.Und)
	return caser.String(s)
}
