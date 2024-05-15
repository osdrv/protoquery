package protoquery

import (
	"reflect"
	"testing"
)

func TestParseQuery(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []*Token
		wantErr error
	}{
		{
			name:  "nodename",
			input: "nodename",
			want:  []*Token{NewToken("nodename", TokenNode)},
		},
		{
			name:  "root node",
			input: "/",
			want:  []*Token{NewToken("/", TokenSlash)},
		},
		{
			name:  "slash slash",
			input: "//",
			want:  []*Token{NewToken("//", TokenSlashSlash)},
		},
		{
			name:  "dot",
			input: ".",
			want:  []*Token{NewToken(".", TokenDot)},
		},
		{
			name:  "dot dot",
			input: "..",
			want:  []*Token{NewToken("..", TokenDotDot)},
		},
		{
			name:  "at",
			input: "@",
			want:  []*Token{NewToken("@", TokenAt)},
		},
		{
			name:  "single quoted string",
			input: "'string'",
			want:  []*Token{NewToken("string", TokenString)},
		},
		{
			name:  "double quoted string",
			input: "\"string\"",
			want:  []*Token{NewToken("string", TokenString)},
		},
		{
			name:  "number",
			input: "123.45",
			want:  []*Token{NewToken("123.45", TokenNumber)},
		},
		{
			name:  "root element",
			input: "/root",
			want: []*Token{
				NewToken("/", TokenSlash),
				NewToken("root", TokenNode),
			},
		},
		{
			name:  "root element with child",
			input: "/root/child",
			want: []*Token{
				NewToken("/", TokenSlash),
				NewToken("root", TokenNode),
				NewToken("/", TokenSlash),
				NewToken("child", TokenNode),
			},
		},
		{
			name:  "all elements",
			input: "//book",
			want: []*Token{
				NewToken("//", TokenSlashSlash),
				NewToken("book", TokenNode),
			},
		},
		{
			name:  "descendant elements",
			input: "book//title",
			want: []*Token{
				NewToken("book", TokenNode),
				NewToken("//", TokenSlashSlash),
				NewToken("title", TokenNode),
			},
		},
		{
			name:  "all attributes",
			input: "//@lang",
			want: []*Token{
				NewToken("//", TokenSlashSlash),
				NewToken("@", TokenAt),
				NewToken("lang", TokenNode),
			},
		},
		{
			name:  "first element",
			input: "/bookstore/book[1]",
			want: []*Token{
				NewToken("/", TokenSlash),
				NewToken("bookstore", TokenNode),
				NewToken("/", TokenSlash),
				NewToken("book", TokenNode),
				NewToken("[", TokenLBracket),
				NewToken("1", TokenNumber),
				NewToken("]", TokenRBracket),
			},
		},
		{
			name:  "last element",
			input: "/bookstore/book[last()]",
			want: []*Token{
				NewToken("/", TokenSlash),
				NewToken("bookstore", TokenNode),
				NewToken("/", TokenSlash),
				NewToken("book", TokenNode),
				NewToken("[", TokenLBracket),
				NewToken("last", TokenNode),
				NewToken("(", TokenLParen),
				NewToken(")", TokenRParen),
				NewToken("]", TokenRBracket),
			},
		},
		{
			name:  "last but one",
			input: "/bookstore/book[last()-1]",
			want: []*Token{
				NewToken("/", TokenSlash),
				NewToken("bookstore", TokenNode),
				NewToken("/", TokenSlash),
				NewToken("book", TokenNode),
				NewToken("[", TokenLBracket),
				NewToken("last", TokenNode),
				NewToken("(", TokenLParen),
				NewToken(")", TokenRParen),
				NewToken("-", TokenMinus),
				NewToken("1", TokenNumber),
				NewToken("]", TokenRBracket),
			},
		},
		{
			name:  "first 2 elements",
			input: "/bookstore/book[position()<3]",
			want: []*Token{
				NewToken("/", TokenSlash),
				NewToken("bookstore", TokenNode),
				NewToken("/", TokenSlash),
				NewToken("book", TokenNode),
				NewToken("[", TokenLBracket),
				NewToken("position", TokenNode),
				NewToken("(", TokenLParen),
				NewToken(")", TokenRParen),
				NewToken("<", TokenLess),
				NewToken("3", TokenNumber),
				NewToken("]", TokenRBracket),
			},
		},
		{
			name:  "elements that have an attribute",
			input: "/book[@lang]",
			want: []*Token{
				NewToken("/", TokenSlash),
				NewToken("book", TokenNode),
				NewToken("[", TokenLBracket),
				NewToken("@", TokenAt),
				NewToken("lang", TokenNode),
				NewToken("]", TokenRBracket),
			},
		},
		{
			name:  "elements that have given attribute value",
			input: "/book[@lang='en']",
			want: []*Token{
				NewToken("/", TokenSlash),
				NewToken("book", TokenNode),
				NewToken("[", TokenLBracket),
				NewToken("@", TokenAt),
				NewToken("lang", TokenNode),
				NewToken("=", TokenEqual),
				NewToken("en", TokenString),
				NewToken("]", TokenRBracket),
			},
		},
		{
			name:  "element have value greather than",
			input: "/book[price>35]",
			want: []*Token{
				NewToken("/", TokenSlash),
				NewToken("book", TokenNode),
				NewToken("[", TokenLBracket),
				NewToken("price", TokenNode),
				NewToken(">", TokenGreater),
				NewToken("35", TokenNumber),
				NewToken("]", TokenRBracket),
			},
		},
		{
			name:  "children of elements with attribute greather than",
			input: "/book[price>35]/title",
			want: []*Token{
				NewToken("/", TokenSlash),
				NewToken("book", TokenNode),
				NewToken("[", TokenLBracket),
				NewToken("price", TokenNode),
				NewToken(">", TokenGreater),
				NewToken("35", TokenNumber),
				NewToken("]", TokenRBracket),
				NewToken("/", TokenSlash),
				NewToken("title", TokenNode),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := parseQuery(tt.input)
			if !errorEqual(tt.wantErr, err) {
				t.Errorf("parseQuery() error = %v, want %v", err, tt.wantErr)
				return
			}
			if tt.wantErr != nil {
				return
			}

			if !reflect.DeepEqual(tt.want, tokens) {
				t.Errorf("Unexpected result: want: %+v, got: %+v", tt.want, tokens)
			}
		})
	}
}
