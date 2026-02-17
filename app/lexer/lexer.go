package lexer


type Lexer struct {
	input []byte
	position int
	readPosition int
	Ch byte
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.Ch = 0
	} else {
		l.Ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

func NewLexer(i []byte) *Lexer {
	l := &Lexer{
		input: i,
	}
	l.readChar()
	return l
}

func (l *Lexer) NextToken() Token {
	var tok Token

	switch l.Ch {
	case '*':
		tok = newToken(ASTERISK, "*")
	case '$':
		tok = newToken(DOLLAR, "$")
	case '\r':
		if l.isTerminator() {
			tok = newToken(TERMINATOR, "\r\n")
			l.readChar()
		} else {
			tok = newToken(RETURN, "\r")
		}
	case '\n':
		tok = newToken(NEWLINE, "\n")
	case 0:
		tok.Type = EOF
		return tok
	default:
		if l.isDigit(l.Ch) {
			tok.Type = INT
			tok.Literal = l.readNumber()
			return tok
		} else if l.isLetter(l.Ch) {
			tok.Literal = l.readString()
			tok.Type = isBuiltinCmd(tok.Literal)
			return tok
		} else {
			tok.Type = ILLEGAL
			return tok
		}
	}
	l.readChar()
	return tok
}

func (l *Lexer) isDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

func (l *Lexer) readNumber() string {
	startP := l.position
	for {
		if !l.isDigit(l.Ch) {
			break
		}
		l.readChar()
	}
	return string(l.input[startP:l.position])
}

func (l *Lexer) isLetter(c byte) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}

func (l *Lexer) readString() string {
	startP := l.position
	for {
		if !l.isLetter(l.Ch) {
			break
		}
		l.readChar()
	}
	return string(l.input[startP:l.position])
}

func (l *Lexer) isTerminator() bool {
	return l.readPosition < len(l.input) && l.input[l.readPosition] == '\n'
}