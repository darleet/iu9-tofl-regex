package parser

type Service struct {
	allowedChars map[rune]struct{}
	digits       map[rune]struct{}
}

func New() *Service {
	return &Service{
		allowedChars: GetAllowedChars(),
		digits:       GetDigits(),
	}
}
