package protoquery

import "fmt"

func isAlpha(s string, ix int) bool {
	return ix < len(s) && (s[ix] >= 'a' && s[ix] <= 'z' || s[ix] >= 'A' && s[ix] <= 'Z')
}

func isDigit(s string, ix int) bool {
	return ix < len(s) && s[ix] >= '0' && s[ix] <= '9'
}

func readNumber(s string, ix int) (string, int) {
	start := ix
	for ix < len(s) && (isDigit(s, ix) || s[ix] == '.') {
		ix++
	}
	return s[start:ix], ix
}

func readNode(s string, ix int) (string, int) {
	start := ix
	for ix < len(s) && isAlpha(s, ix) {
		ix++
	}
	return s[start:ix], ix
}

func match(s string, ix int, ch TokenKind) bool {
	return ix < len(s) && s[ix] == byte(ch)
}

func matchAny(s string, ix int, chs ...TokenKind) bool {
	for _, ch := range chs {
		if match(s, ix, ch) {
			return true
		}
	}
	return false
}

func isWhitespace(s string, ix int) bool {
	return ix < len(s) && s[ix] == ' '
}

func eatWhitespace(s string, ix int) int {
	for isWhitespace(s, ix) {
		ix++
	}
	return ix
}

func parseQuery(query string) ([]*Token, error) {
	tokens := make([]*Token, 0, 1)
	ix := 0
	for ix < len(query) {
		start := ix
		if isWhitespace(query, ix) {
			ix = eatWhitespace(query, ix)
		} else if match(query, ix, TokenSlash) {
			tk := TokenSlash
			if len(query) > ix && match(query, ix+1, TokenSlash) {
				tk = TokenSlashSlash
				ix++
			}
			ix++
			tokens = append(tokens, NewToken(query[start:ix], tk))
		} else if match(query, ix, TokenDot) {
			tk := TokenDot
			if len(query) > ix && match(query, ix+1, TokenDot) {
				tk = TokenDotDot
				ix++
			}
			ix++
			tokens = append(tokens, NewToken(query[start:ix], tk))
		} else if match(query, ix, TokenAt) {
			ix++
			tokens = append(tokens, NewToken(query[start:ix], TokenAt))
		} else if matchAny(query, ix, TokenLBracket, TokenRBracket, TokenLParen,
			TokenRParen, TokenStar, TokenEqual, TokenLess, TokenGreater, TokenMinus, TokenPipe) {
			tokens = append(tokens, NewToken(query[ix:ix+1], TokenKind(query[ix])))
			ix++
		} else if matchAny(query, ix, TokenSingleQuote, TokenDoubleQuote) {
			end := query[ix] // we're looking for the matching quote
			ix++
			for ix < len(query) && query[ix] != end {
				ix++
			}
			if query[ix] != end {
				return nil, fmt.Errorf("unterminated string at position %d", ix)
			}
			tokens = append(tokens, NewToken(query[start+1:ix], TokenString))
			ix++
		} else if isAlpha(query, ix) {
			var node string
			node, ix = readNode(query, ix)
			tokens = append(tokens, NewToken(node, TokenNode))
		} else if isDigit(query, ix) {
			var number string
			number, ix = readNumber(query, ix)
			tokens = append(tokens, NewToken(number, TokenNumber))
		} else {
			return nil, fmt.Errorf("unexpected character %q at position %d", query[ix], ix)
		}
	}
	return tokens, nil
}
