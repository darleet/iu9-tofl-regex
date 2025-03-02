package parser_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/darleet/iu9-tofl-regex/internal/service/parser"
)

func TestService_Parse_Valid(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		// -- valid
		{"одиночный_символ", "a", false},
		{"символ_в_группе", "(aa)", false},
		{"символы_вокруг_группы", "aa(bb)aa", false},
		{"астериск", "a*", false},
		{"группа_с_астериском", "(ab)*", false},
		{"альтернатива_с_астериском", "(aa|bb)*", false},
		{"ссылка_на_выражение", "(aa|bb)(?1)", false},
		{"ссылка_на_выражение_изнутри", "(a(?1)b)", false},
		{"ссылка_на_выражение_изнутри_с_альтернативой", "(aa|b(?1))", false},
		{"ссылка_на_альтернативу_со_вложенностью", "(a|(bb))(?2)", false},
		{"множественная_альтернатива", "(aa|bb|cc|dd(?1))(?1)", false},
		{"ссылка_на_строку", "(aa|bb)\\1", false},
		{"ссылка_на_строку_3", "(a|(bb))(a|b)\\3", false},
		{"групп_захвата_9", "(((((((((a)a)a)a)a)a)a)a)a)", false},
		{"ссылка_на_выражение", "(a|(bb))(a|(?3))", false},
		{"ссылка_на_выражение", "(a(?1)b|c)", false},
		{"ссылка_на_выражение", "(a(?1)a|b)", false},
		{"ссылка_на_выражение", "((?1)a|b)", false},
		{"ссылка_на_выражение", "(?1)(a|b)*(?1)", false},
		{"незахватывающая_группа", "((?:a(?2)|(bb))(?1))", false},
		// -- invalid
		{"некорректный_ввод", "INVALID INPUT", true},
		{"ссылка_на_несуществующую_строку", "(aa|bb)\\2", true},
		{"ссылка_на_неинициализированную_строку", "(\\1)(a|b)", true},
		{"незакрытые_скобки", "((aa)abaa", true},
		{"некоткрытые_скобки", "(aa)b))a)", true},
		{"групп_захвата_больше_9", "((((((((((a)a)a)a)a)a)a)a)a)a)", true},
		{"ссылка_на_ветку_альтернативы", "(a|(bb))(a|\\2)", true},
		{"ссылка_на_группу_с_ссылкой_на_выражение", "(a|(?2))(a|(bb\\1))", true},
		{"ссылка_на_альтернативу_с_астериском", "(a|b)*\\1", true},
		{"ссылка_на_группу_ветки_альтернативы", "(a(bb)|b(cc))\\2", true},
	}

	for _, tt := range tests {
		var name string
		if tt.wantErr {
			name = "invalid/" + tt.name
		} else {
			name = "valid/" + tt.name
		}
		t.Run(name, func(t *testing.T) {
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
