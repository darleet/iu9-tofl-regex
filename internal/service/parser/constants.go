package parser

func GetOperators() map[rune]struct{} {
	m := make(map[rune]struct{})
	for _, r := range "()|*?:\\" {
		m[r] = struct{}{}
	}
	return m
}

func GetLetters() map[rune]struct{} {
	m := make(map[rune]struct{})
	for r := 'a'; r <= 'z'; r++ {
		m[r] = struct{}{}
	}
	return m
}

func GetDigits() map[rune]struct{} {
	m := make(map[rune]struct{})
	for r := '1'; r <= '9'; r++ {
		m[r] = struct{}{}
	}
	return m
}

func GetAllowedChars() map[rune]struct{} {
	m := make(map[rune]struct{})
	for r := range GetOperators() {
		m[r] = struct{}{}
	}
	for r := range GetLetters() {
		m[r] = struct{}{}
	}
	for r := range GetDigits() {
		m[r] = struct{}{}
	}
	return m
}
