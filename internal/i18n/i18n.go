package i18n

import "strings"

type Language string

const (
	English Language = "en"
	Spanish Language = "es"
)

func ParseLanguage(raw string) (Language, bool) {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "en", "english", "inglés", "ingles":
		return English, true
	case "es", "spanish", "español", "espanol":
		return Spanish, true
	default:
		return "", false
	}
}

func NormalizeLanguage(raw string) Language {
	if lang, ok := ParseLanguage(raw); ok {
		return lang
	}
	return English
}

func IsSupported(raw string) bool {
	_, ok := ParseLanguage(raw)
	return ok
}

func T(lang Language, en, es string) string {
	if lang == Spanish {
		return es
	}
	return en
}
