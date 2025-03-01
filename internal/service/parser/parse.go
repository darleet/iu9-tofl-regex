package parser

import "context"

func (s *Service) Parse(ctx context.Context, regex string) bool {
	r := []rune(regex)

	var i int

	for i < len(r) {
		_, ok := s.allowedChars[r[i]]
		if !ok {
			return false
		}
		i++
	}

	return true
}
