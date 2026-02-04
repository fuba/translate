package chunk

func Split(text string, maxChars int) []string {
	if maxChars <= 0 {
		return []string{text}
	}
	if text == "" {
		return []string{""}
	}

	runes := []rune(text)
	if len(runes) <= maxChars {
		return []string{text}
	}

	chunks := make([]string, 0, (len(runes)/maxChars)+1)
	start := 0
	for start < len(runes) {
		end := start + maxChars
		if end > len(runes) {
			end = len(runes)
		}

		cut := findBoundary(runes, start, end)
		if cut <= start {
			cut = end
		}
		chunks = append(chunks, string(runes[start:cut]))
		start = cut
	}
	return chunks
}

func findBoundary(runes []rune, start, end int) int {
	for i := end - 1; i > start; i-- {
		if isBoundary(runes[i]) {
			return i + 1
		}
	}
	return -1
}

func isBoundary(r rune) bool {
	switch r {
	case '\n', '。', '．', '！', '？', '!', '?', '.', ',', '、', ';', '；', ':', '：':
		return true
	case ' ', '\t':
		return true
	default:
		return false
	}
}
