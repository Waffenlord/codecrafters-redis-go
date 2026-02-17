package lexer

import (
	"testing"
)



func TestLexer(t *testing.T) {
	testInput := []byte("*2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n")
	l := NewLexer(testInput)
	tok := l.NextToken()
	if tok.Type != ASTERISK {
		t.Error("Token should be *")
	}
	tok = l.NextToken()
	if tok.Type != INT || tok.Literal != "2" {
		t.Error("Token should be INT 2")
	}
	tok = l.NextToken()
	if tok.Type != TERMINATOR {
		t.Error("Token should be TERMINATOR")
	}
	tok = l.NextToken()
	if tok.Type != DOLLAR {
		t.Error("Token should be DOLLAR")
	}
	tok = l.NextToken()
	if tok.Type != INT || tok.Literal != "4" {
		t.Error("Token should be INT 4")
	}
	tok = l.NextToken()
	if tok.Type != TERMINATOR {
		t.Error("Token should be TERMINATOR")
	}
	tok = l.NextToken()
	if tok.Type != CMD && tok.Literal != "ECHO" {
		t.Error("Token should be CMD ECHO")
	}
	tok = l.NextToken()
	if tok.Type != TERMINATOR {
		t.Error("Token should be TERMINATOR")
	}
	tok = l.NextToken()
	if tok.Type != DOLLAR {
		t.Error("Token should be DOLLAR")
	}
	tok = l.NextToken()
	if tok.Type != INT || tok.Literal != "3" {
		t.Error("Token should be INT 3")
	}
	tok = l.NextToken()
	if tok.Type != TERMINATOR {
		t.Error("Token should be TERMINATOR")
	}
	tok = l.NextToken()
	if tok.Type != STRING || tok.Literal != "hey" {
		t.Error("Token should be STRING hey")
	}
	tok = l.NextToken()
	if tok.Type != TERMINATOR {
		t.Error("Token should be TERMINATOR")
	}
	tok = l.NextToken()
	if tok.Type != EOF {
		t.Error("Token should be EOF")
	}
}