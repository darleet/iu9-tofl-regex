package parser_test

import (
	"context"
	"testing"

	"github.com/darleet/iu9-tofl-regex/internal/service/parser"
)

func TestService_Parse(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"valid/ссылка_на_выражение", "(aa|bb)(?1)", true},
		{"valid/ссылка_на_строку", "(aa|bb)\\1", true},
		{"invalid/некорректный_ввод", "INVALID INPUT", false},
		{"invalid/ссылка_на_несуществующую_строку", "(aa|bb)\\2", false},
		{"invalid/ссылка_на_неинициализированную_строку", "(\\1)(a|b)", false},
		{"invalid/незакрытые_скобки", "((aa)abaa", false},
		{"invalid/неоткрытые_скобки", "aa)aaab", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := parser.New()
			if got := s.Parse(context.Background(), tt.in); got != tt.want {
				t.Errorf("Service.Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
