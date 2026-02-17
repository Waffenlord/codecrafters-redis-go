package parser

import (
	"errors"
	"strconv"
	"github.com/codecrafters-io/redis-starter-go/app/lexer"
)


type Parser struct {
	l *lexer.Lexer
	curToken lexer.Token
	peekToken lexer.Token
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func New(lx *lexer.Lexer) *Parser {
	p := &Parser{
		l: lx,
	}

	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) ParseProgram() (Node, error) {
	switch p.curToken.Type {
	case lexer.ASTERISK:
		return p.parseArray()
	case lexer.DOLLAR:
		return p.parseBulkString()
	}
	return nil, nil
}

func (p *Parser) parseArray() (Node, error) {
	arr := Array{
		Tok: p.curToken,
		Elements: make([]Node, 0),
	}
	if p.peekToken.Type != lexer.INT {
		return nil, errors.New("missing length of the array")
	}
	p.nextToken()
	var err error
	arr.Length, err = strconv.Atoi(p.curToken.Literal)
	if err != nil {
		return nil, errors.New("error occurred transforming the array length")
	}
	if p.peekToken.Type != lexer.TERMINATOR {
		return nil, errors.New("missing separator of the array")
	}
	p.nextToken()
	p.nextToken()
	for p.curToken.Type != lexer.EOF {
		element, err := p.ParseProgram()
		if err != nil {
			return nil, errors.New("error parsing array element")
		}
		arr.Elements = append(arr.Elements, element)
	}

	if arr.Length != len(arr.Elements) {
		return nil, errors.New("incorrect number of elements in the array")
	}

	return arr, nil
}

func (p *Parser) parseBulkString() (Node, error) {
	if p.peekToken.Type != lexer.INT {
		return nil, errors.New("missing length of the string")
	}
	p.nextToken()
	length, err := strconv.Atoi(p.curToken.Literal)
	if err != nil {
		return nil, errors.New("error occurred transforming the string length")
	}
	s := BulkString{
		Length: length,
	}
	if p.peekToken.Type != lexer.TERMINATOR {
		return nil, errors.New("missing separator of the array")
	}
	p.nextToken()
	if p.peekToken.Type != lexer.CMD && p.peekToken.Type != lexer.STRING {
		return nil, errors.New("invalid value for string")
	}
	p.nextToken()
	s.Tok = p.curToken
	s.Literal = p.curToken.Literal
	p.nextToken()
	p.nextToken()

	return s, nil
}