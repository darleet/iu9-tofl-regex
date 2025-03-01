package parser

func GetAllowedChars() map[rune]struct{} {
	m := make(map[rune]struct{})

	for _, r := range "()|*?:\\" {
		m[r] = struct{}{}
	}

	for r := 'a'; r <= 'z'; r++ {
		m[r] = struct{}{}
	}

	for r := '1'; r <= '9'; r++ {
		m[r] = struct{}{}
	}

	return m
}
