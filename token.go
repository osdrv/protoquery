package protoquery

import (
	"fmt"
	"strconv"
)

type TokenKind byte

const (
	TokenNone         TokenKind = iota
	TokenAnd          TokenKind = '&' // And is a pseudo-token that represents a logical AND operator.
	TokenAt           TokenKind = '@'
	TokenBang         TokenKind = '!'
	TokenBool         TokenKind = 'B' // Bool is a pseudo-token that represents a boolean.
	TokenComma        TokenKind = ','
	TokenDot          TokenKind = '.'
	TokenDotDot       TokenKind = ':' // DotDot is a pseudo-token that represents a double dot.
	TokenDoubleQuote  TokenKind = '"'
	TokenEqual        TokenKind = '='
	TokenNotEqual     TokenKind = '!' // NotEqual is a pseudo-token that represents a not equal operator.
	TokenGreater      TokenKind = '>'
	TokenGreaterEqual TokenKind = 'G' // GreaterEqual is a pseudo-token that represents a greater than or equal operator.
	TokenLBracket     TokenKind = '['
	TokenLParen       TokenKind = '('
	TokenLess         TokenKind = '<'
	TokenLessEqual    TokenKind = 'L' // LessEqual is a pseudo-token that represents a less than or equal operator.
	TokenMinus        TokenKind = '-'
	TokenNode         TokenKind = 'N' // Node is a pseudo-token that represents a node.
	TokenNumber       TokenKind = '0' // Number is a pseudo-token that represents a number.
	TokenOr           TokenKind = 'O' // Or is a pseudo-token that represents a logical OR operator.
	TokenPipe         TokenKind = '|'
	TokenPlus         TokenKind = '+'
	TokenRBracket     TokenKind = ']'
	TokenRParen       TokenKind = ')'
	TokenSingleQuote  TokenKind = '\''
	TokenSlash        TokenKind = '/'
	TokenSlashSlash   TokenKind = '\\' // SlashSlash is a pseudo-token that represents a double slash.
	TokenStar         TokenKind = '*'
	TokenString       TokenKind = 'S' // String is a pseudo-token that represents a string.
)

type Token struct {
	Kind  TokenKind
	Value string
}

func NewToken(value string, kind TokenKind) *Token {
	return &Token{Value: value, Kind: kind}
}

func (t *Token) String() string {
	switch t.Kind {
	case TokenString:
		return fmt.Sprintf("'%s'", t.Value)
	default:
		return t.Value
	}
}

func (t *Token) IntValue() (int64, error) {
	if t.Kind != TokenNumber {
		return 0, fmt.Errorf("Token is not a number: %v", t.Kind)
	}
	ix, err := strconv.ParseInt(t.Value, 10, 64)
	if err != nil {
		return 0, err
	}
	return ix, nil
}

func (t *Token) BoolValue() (bool, error) {
	if t.Kind != TokenBool {
		return false, fmt.Errorf("Token is not a boolean: %v", t.Kind)
	}
	return strconv.ParseBool(t.Value)
}
