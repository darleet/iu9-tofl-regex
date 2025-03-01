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
		{"valid", "(aa|bb)(?1)", true},
		{"valid", "(aa|bb)\\1", true},
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
