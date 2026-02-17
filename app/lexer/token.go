package lexer

import "strings"

type TokenType string

type Token struct {
	Type TokenType
	Literal string
}


const (
	// Special characters
	ASTERISK = "*"
	DOLLAR = "$"
	COLON = ":"
	PLUS = "+"
	RETURN = "\r"
	NEWLINE = "\n"

	// Primitives
	INT = "INT"
	STRING = "STRING"

	// Special categories
	CMD = "CMD"
	TERMINATOR = "TERMINATOR"


	ILLEGAL = "ILLEGAL"
	EOF = "EOF"
)


var builtinKeywords = map[string]TokenType{
	"echo": CMD,
	"ping": CMD,
}

func isBuiltinCmd(s string) TokenType {
	lowerS := strings.ToLower(s)
	if t, ok := builtinKeywords[lowerS]; ok {
		return t
	}

	return STRING
}


func newToken(tokenType TokenType, literal string) Token {
	return Token{Type: tokenType, Literal: literal}
}
