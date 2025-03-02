package parser_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/darleet/iu9-tofl-regex/internal/service/parser"
)

func TestService_Parse(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{"valid/ссылка_на_выражение", "(aa|bb)(?1)", false},
		{"valid/ссылка_на_строку", "(aa|bb)\\1", false},
		{"valid/групп_захвата_9", "(((((((((a)a)a)a)a)a)a)a)a)", false},
		{"invalid/некорректный_ввод", "INVALID INPUT", true},
		{"invalid/ссылка_на_несуществующую_строку", "(aa|bb)\\2", true},
		{"invalid/ссылка_на_неинициализированную_строку", "(\\1)(a|b)", true},
		{"invalid/незакрытые_скобки", "((aa)abaa", true},
		{"invalid/неоткрытые_скобки", "aa)aaab", true},
		{"invalid/групп_захвата_больше_9", "((((((((((a)a)a)a)a)a)a)a)a)a)", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := parser.New()
			got, err := s.Parse(context.Background(), tt.in)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}
