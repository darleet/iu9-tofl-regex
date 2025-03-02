package parser

type Service struct {
	allowedChars map[rune]struct{}
	operators    map[rune]struct{}
	letters      map[rune]struct{}
	digits       map[rune]struct{}
}

func New() *Service {
	return &Service{
		allowedChars: GetAllowedChars(),
		operators:    GetOperators(),
		letters:      GetLetters(),
		digits:       GetDigits(),
	}
}
