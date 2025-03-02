package parser

import "context"

func (s *Service) Parse(ctx context.Context, regex string) bool {
	r := []rune(regex)

	var i int
	var brCount, grCount int
	var maxNum rune

	for i < len(r) {
		_, ok := s.allowedChars[r[i]]
		if !ok {
			return false
		}

		if i < len(r)-3 && r[i] == '(' && r[i+1] == '?' && s.IsDigit(r[i+2]) && r[i+3] == ')' {
			if r[i+2] > maxNum {
				maxNum = r[i+2]
			}
			i += 4
		} else if i < len(r)-3 && r[i] == '(' && r[i+1] == '\\' && s.IsDigit(r[i+2]) && r[i+3] == ')' {
			if int(r[i+2]-'0') > grCount {
				return false
			}
			i += 4
		} else if i < len(r)-1 && r[i] == '\\' && s.IsDigit(r[i+1]) {
			if int(r[i+1]-'0') > grCount {
				return false
			}
			i += 2
		} else if r[i] == '(' {
			brCount++
		} else if r[i] == ')' {
			brCount--
			grCount++
		}

		i++
	}

	if brCount != 0 || grCount > 9 || int(maxNum-'0') > grCount {
		return false
	}

	return true
}
