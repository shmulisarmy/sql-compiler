package tokenizer

import (
	"unicode"
)

// ---------------- TOKEN TYPES ----------------

type TokenType string

const (
	// Keywords
	SELECT TokenType = "SELECT"
	FROM   TokenType = "FROM"
	WHERE  TokenType = "WHERE"
	AND    TokenType = "AND"
	TRUE   TokenType = "TRUE"
	FALSE  TokenType = "FALSE"
	AS     TokenType = "AS"

	// Special
	ILLEGAL TokenType = "ILLEGAL"
	EOF     TokenType = "EOF"

	// Identifiers + literals
	IDENT  TokenType = "IDENT"
	INT    TokenType = "INT"
	FLOAT  TokenType = "FLOAT"
	STRING TokenType = "STRING"

	// Operators
	ASSIGN   TokenType = "="
	PLUS     TokenType = "+"
	MINUS    TokenType = "-"
	BANG     TokenType = "!"
	ASTERISK TokenType = "*"
	SLASH    TokenType = "/"

	LT TokenType = "<"
	LE TokenType = "<="
	GT TokenType = ">"
	GE TokenType = ">="
	EQ TokenType = "=="

	// Delimiters
	COMMA     TokenType = ","
	SEMICOLON TokenType = ";"
	COLON     TokenType = ":"
	DOT       TokenType = "."
	LPAREN    TokenType = "("
	RPAREN    TokenType = ")"
	LBRACE    TokenType = "{"
	RBRACE    TokenType = "}"
)

// ---------------- TOKEN STRUCT ----------------

type Token struct {
	Type    TokenType
	Literal string
	Pos     int
}

// ---------------- LEXER ----------------

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           rune
}

func NewLexer(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		r := rune(l.input[l.readPosition])
		size := 1
		if r >= 0x80 {
			r, size = utf8DecodeRuneInStringAt(l.input, l.readPosition)
		}
		l.ch = r
		l.readPosition += size
	}
	l.position = l.readPosition - 1
}

func utf8DecodeRuneInStringAt(s string, i int) (r rune, size int) {
	return rune(s[i]), 1
}

func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.input) {
		return 0
	}
	r, _ := utf8DecodeRuneInStringAt(l.input, l.readPosition)
	return r
}

// ---------------- KEYWORDS ----------------

var keywords = map[string]TokenType{
	"SELECT": SELECT,
	"select": SELECT,
	"FROM":   FROM,
	"from":   FROM,
	"WHERE":  WHERE,
	"where":  WHERE,
	"AND":    AND,
	"and":    AND,
	"true":   TRUE,
	"false":  FALSE,
	"AS":     AS,
	"as":     AS,
}

func lookupIdent(ident string) TokenType {
	// SQL-style: treat keywords as uppercase only
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

// ---------------- MAIN TOKENIZER ----------------

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	tok := Token{Pos: l.position}

	switch l.ch {
	case '"':
		tok.Type = STRING
		tok.Literal = l.readString()
		return tok

	case ',':
		tok = newToken(COMMA, l.ch, l.position)
	case ';':
		tok = newToken(SEMICOLON, l.ch, l.position)
	case ':':
		tok = newToken(COLON, l.ch, l.position)
	case '(':
		tok = newToken(LPAREN, l.ch, l.position)
	case ')':
		tok = newToken(RPAREN, l.ch, l.position)
	case '{':
		tok = newToken(LBRACE, l.ch, l.position)
	case '}':
		tok = newToken(RBRACE, l.ch, l.position)
	case '.':
		tok = newToken(DOT, l.ch, l.position)
	case '+':
		tok = newToken(PLUS, l.ch, l.position)
	case '-':
		tok = newToken(MINUS, l.ch, l.position)
	case '*':
		tok = newToken(ASTERISK, l.ch, l.position)
	case '/':
		if l.peekChar() == '/' {
			l.readChar()
			l.readChar()
			for l.ch != '\n' && l.ch != 0 {
				l.readChar()
			}
			return l.NextToken()
		}
		tok = newToken(SLASH, l.ch, l.position)
	case '!':
		tok = newToken(BANG, l.ch, l.position)
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok = newToken(LE, l.ch, l.position)
		} else {
			tok = newToken(LT, l.ch, l.position)
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = newToken(GE, l.ch, l.position)
		} else {
			tok = newToken(GT, l.ch, l.position)
		}
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			tok = newToken(EQ, l.ch, l.position)
		} else {
			tok = newToken(ASSIGN, l.ch, l.position)
		}
	case 0:
		tok.Type = EOF
		tok.Literal = ""
	default:
		if isLetter(l.ch) {
			start := l.position
			lit := l.readIdentifier()
			return Token{
				Type:    lookupIdent(lit),
				Literal: lit,
				Pos:     start,
			}
		} else if isDigit(l.ch) {
			start := l.position
			lit, ttype := l.readNumber()
			return Token{
				Type:    ttype,
				Literal: lit,
				Pos:     start,
			}
		} else {
			tok = newToken(ILLEGAL, l.ch, l.position)
		}
	}

	l.readChar()
	return tok
}

func newToken(tt TokenType, ch rune, pos int) Token {
	return Token{Type: tt, Literal: string(ch), Pos: pos}
}

func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(l.ch) {
		l.readChar()
	}
}

// ---------------- IDENTIFIERS ----------------

func isLetter(ch rune) bool {
	return ch == '_' || unicode.IsLetter(ch)
}

func (l *Lexer) readIdentifier() string {
	start := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[start:l.position]
}

// ---------------- NUMBERS ----------------

func isDigit(ch rune) bool {
	return unicode.IsDigit(ch)
}

func (l *Lexer) readNumber() (string, TokenType) {
	start := l.position
	tt := INT
	hasDot := false

	for isDigit(l.ch) || (!hasDot && l.ch == '.') {
		if l.ch == '.' {
			hasDot = true
			tt = FLOAT
		}
		l.readChar()
	}

	return l.input[start:l.position], tt
}

// ---------------- STRINGS ----------------

func (l *Lexer) readString() string {
	l.readChar()
	escaped := false
	result := []rune{}

	for l.ch != 0 {
		if l.ch == '"' && !escaped {
			l.readChar()
			break
		}
		if l.ch == '\\' && !escaped {
			escaped = true
			l.readChar()
			continue
		}
		if escaped {
			switch l.ch {
			case 'n':
				result = append(result, '\n')
			case 't':
				result = append(result, '\t')
			case '"':
				result = append(result, '"')
			case '\\':
				result = append(result, '\\')
			default:
				result = append(result, l.ch)
			}
			escaped = false
		} else {
			result = append(result, l.ch)
		}
		l.readChar()
	}

	return string(result)
}

func (l *Lexer) Tokenize() []Token {
	tokens := []Token{}
	for l.ch != 0 {
		tokens = append(tokens, l.NextToken())
	}
	return tokens
}
