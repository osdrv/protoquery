package protoquery

import "fmt"

type TokenKind byte

type Token struct {
	Kind  TokenKind
	Value string
}

func (t *Token) String() string {
	switch t.Kind {
	case TokenString:
		return fmt.Sprintf("'%s'", t.Value)
	default:
		return t.Value
	}
}

const (
	TokenNone         TokenKind = iota
	TokenAt           TokenKind = '@'
	TokenDot          TokenKind = '.'
	TokenDotDot       TokenKind = ':' // DotDot is a pseudo-token that represents a double dot.
	TokenDoubleQuote  TokenKind = '"'
	TokenEqual        TokenKind = '='
	TokenGreater      TokenKind = '>'
	TokenGreaterEqual TokenKind = 'G'
	TokenLBracket     TokenKind = '['
	TokenLParen       TokenKind = '('
	TokenLess         TokenKind = '<'
	TokenLessEqual    TokenKind = 'L'
	TokenMinus        TokenKind = '-'
	TokenNode         TokenKind = 'N'
	TokenNumber       TokenKind = '0'
	TokenPipe         TokenKind = '|'
	TokenRBracket     TokenKind = ']'
	TokenRParen       TokenKind = ')'
	TokenSingleQuote  TokenKind = '\''
	TokenSlash        TokenKind = '/'
	TokenSlashSlash   TokenKind = '\\' // SlashSlash is a pseudo-token that represents a double slash.
	TokenStar         TokenKind = '*'
	TokenString       TokenKind = 'S'
)

func NewToken(value string, kind TokenKind) *Token {
	return &Token{Value: value, Kind: kind}
}
