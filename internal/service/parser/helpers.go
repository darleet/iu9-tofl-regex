package parser

func (s *Service) IsOperator(r rune) bool {
	_, ok := s.operators[r]
	return ok
}

func (s *Service) IsLetter(r rune) bool {
	_, ok := s.letters[r]
	return ok
}

func (s *Service) IsDigit(r rune) bool {
	_, ok := s.digits[r]
	return ok
}
