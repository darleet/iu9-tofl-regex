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
		{"valid/одиночный_символ", "a", false},
		{"valid/символ_в_группе", "(aa)", false},
		{"valid/символы_вокруг_группы", "aa(bb)aa", false},
		{"valid/астериск", "a*", false},
		{"valid/группа_с_астериском", "(ab)*", false},
		{"valid/альтернатива_с_астериском", "(aa|bb)*", false},
		{"valid/ссылка_на_выражение", "(aa|bb)(?1)", false},
		{"valid/ссылка_на_выражение_изнутри", "(a(?1)b)", false},
		{"valid/ссылка_на_выражение_изнутри_с_альтернативой", "(aa|b(?1))", false},
		{"valid/ссылка_на_альтернативу_со_вложенностью", "(a|(bb))(?2)", false},
		{"valid/множественная_альтернатива", "(aa|bb|cc|dd(?1))(?1)", false},
		{"valid/ссылка_на_строку", "(aa|bb)\\1", false},
		{"valid/ссылка_на_строку_3", "(a|(bb))(a|b)\\3", false},
		{"valid/групп_захвата_9", "(((((((((a)a)a)a)a)a)a)a)a)", false},
		{"invalid/некорректный_ввод", "INVALID INPUT", true},
		{"invalid/ссылка_на_несуществующую_строку", "(aa|bb)\\2", true},
		{"invalid/ссылка_на_неинициализированную_строку", "(\\1)(a|b)", true},
		{"invalid/незакрытые_скобки", "((aa)abaa", true},
		{"invalid/некоткрытые_скобки", "(aa)b))a)", true},
		{"invalid/групп_захвата_больше_9", "((((((((((a)a)a)a)a)a)a)a)a)a)", true},
		{"invalid/ссылка_на_ветку_альтернативы", "(a|(bb))(a|\\2)", true},
		{"invalid/ссылка_на_группу_с_ссылкой_на_выражение", "(a|(?2))(a|(bb\\1))", true},
		{"invalid/ссылка_на_альтернативу_с_астериском", "(a|b)*\\1", true},
		{"invalid/ссылка_на_группу_ветки_альтернативы", "(a(bb)|b(cc))\\2", true},
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
