package lang

import "strings"

const fallbackLang = "en"

func DefaultTargetLang(env string) string {
	trimmed := strings.TrimSpace(env)
	if trimmed == "" {
		return fallbackLang
	}

	upper := strings.ToUpper(trimmed)
	if upper == "C" || upper == "POSIX" {
		return fallbackLang
	}

	if i := strings.Index(trimmed, "."); i >= 0 {
		trimmed = trimmed[:i]
	}
	if i := strings.IndexAny(trimmed, "_-"); i >= 0 {
		trimmed = trimmed[:i]
	}

	trimmed = strings.ToLower(strings.TrimSpace(trimmed))
	if trimmed == "" {
		return fallbackLang
	}
	return trimmed
}
