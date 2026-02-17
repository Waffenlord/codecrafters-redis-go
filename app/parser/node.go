package parser

import "github.com/codecrafters-io/redis-starter-go/app/lexer"

type Node interface {
	GetToken() lexer.Token
}

type Array struct {
	Tok lexer.Token
	Elements []Node
	Length int
}

func(a Array) GetToken() lexer.Token {
	return a.Tok
}

type BulkString struct {
	Tok lexer.Token
	Literal string
	Length int
}

func(b BulkString) GetToken() lexer.Token {
	return b.Tok
}
