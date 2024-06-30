package protoquery

import (
	"testing"
)

func TestTokenizeXPathQuery(t *testing.T) {
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
			want:  []*Token{NewToken("123.45", TokenFloat)},
		},
		{
			name:  "bool true",
			input: "True",
			want:  []*Token{NewToken("True", TokenBool)},
		},
		{
			name:  "bool false",
			input: "FALSE",
			want:  []*Token{NewToken("FALSE", TokenBool)},
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
				NewToken("1", TokenInt),
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
				NewToken("1", TokenInt),
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
				NewToken("3", TokenInt),
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
				NewToken("35", TokenInt),
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
				NewToken("35", TokenInt),
				NewToken("]", TokenRBracket),
				NewToken("/", TokenSlash),
				NewToken("title", TokenNode),
			},
		},
		{
			name:  "any element node",
			input: "/bookstore/*",
			want: []*Token{
				NewToken("/", TokenSlash),
				NewToken("bookstore", TokenNode),
				NewToken("/", TokenSlash),
				NewToken("*", TokenStar),
			},
		},
		{
			name:  "any attribute node",
			input: "/bookstore/@*",
			want: []*Token{
				NewToken("/", TokenSlash),
				NewToken("bookstore", TokenNode),
				NewToken("/", TokenSlash),
				NewToken("@", TokenAt),
				NewToken("*", TokenStar),
			},
		},
		{
			name:  "any node of any kind",
			input: "/bookstore/node()",
			want: []*Token{
				NewToken("/", TokenSlash),
				NewToken("bookstore", TokenNode),
				NewToken("/", TokenSlash),
				NewToken("node", TokenNode),
				NewToken("(", TokenLParen),
				NewToken(")", TokenRParen),
			},
		},
		{
			name:  "all child element nodes",
			input: "/bookstore/*",
			want: []*Token{
				NewToken("/", TokenSlash),
				NewToken("bookstore", TokenNode),
				NewToken("/", TokenSlash),
				NewToken("*", TokenStar),
			},
		},
		{
			name:  "all elements in the message",
			input: "//*",
			want: []*Token{
				NewToken("//", TokenSlashSlash),
				NewToken("*", TokenStar),
			},
		},
		{
			name:  "all elements with at least one attribute of any kind",
			input: "//title[@*]",
			want: []*Token{
				NewToken("//", TokenSlashSlash),
				NewToken("title", TokenNode),
				NewToken("[", TokenLBracket),
				NewToken("@", TokenAt),
				NewToken("*", TokenStar),
				NewToken("]", TokenRBracket),
			},
		},
		{
			name:  "several paths",
			input: "/book/title | /book/price",
			want: []*Token{
				NewToken("/", TokenSlash),
				NewToken("book", TokenNode),
				NewToken("/", TokenSlash),
				NewToken("title", TokenNode),
				NewToken("|", TokenPipe),
				NewToken("/", TokenSlash),
				NewToken("book", TokenNode),
				NewToken("/", TokenSlash),
				NewToken("price", TokenNode),
			},
		},
		{
			name:  "nested path withan  attribute filter",
			input: "/bookstore/book[@price>35.00]",
			want: []*Token{
				NewToken("/", TokenSlash),
				NewToken("bookstore", TokenNode),
				NewToken("/", TokenSlash),
				NewToken("book", TokenNode),
				NewToken("[", TokenLBracket),
				NewToken("@", TokenAt),
				NewToken("price", TokenNode),
				NewToken(">", TokenGreater),
				NewToken("35.00", TokenFloat),
				NewToken("]", TokenRBracket),
			},
		},
		{
			name:  "node with an attribute and an index dereferencing",
			input: "/bookstore/book[1][@price>35.00]",
			want: []*Token{
				NewToken("/", TokenSlash),
				NewToken("bookstore", TokenNode),
				NewToken("/", TokenSlash),
				NewToken("book", TokenNode),
				NewToken("[", TokenLBracket),
				NewToken("1", TokenInt),
				NewToken("]", TokenRBracket),
				NewToken("[", TokenLBracket),
				NewToken("@", TokenAt),
				NewToken("price", TokenNode),
				NewToken(">", TokenGreater),
				NewToken("35.00", TokenFloat),
				NewToken("]", TokenRBracket),
			},
		},
		{
			name:  "node with an attribute inequality",
			input: "/bookstore/book[@price!=35.00]",
			want: []*Token{
				NewToken("/", TokenSlash),
				NewToken("bookstore", TokenNode),
				NewToken("/", TokenSlash),
				NewToken("book", TokenNode),
				NewToken("[", TokenLBracket),
				NewToken("@", TokenAt),
				NewToken("price", TokenNode),
				NewToken("!=", TokenNotEqual),
				NewToken("35.00", TokenFloat),
				NewToken("]", TokenRBracket),
			},
		},
		{
			name:  "node with underscore",
			input: "/bookstore/book/some_attribute",
			want: []*Token{
				NewToken("/", TokenSlash),
				NewToken("bookstore", TokenNode),
				NewToken("/", TokenSlash),
				NewToken("book", TokenNode),
				NewToken("/", TokenSlash),
				NewToken("some_attribute", TokenNode),
			},
		},
		{
			name:  "node with digits",
			input: "/bookstore/book/int32",
			want: []*Token{
				NewToken("/", TokenSlash),
				NewToken("bookstore", TokenNode),
				NewToken("/", TokenSlash),
				NewToken("book", TokenNode),
				NewToken("/", TokenSlash),
				NewToken("int32", TokenNode),
			},
		},
		{
			name:  "node with an function call in the attribute",
			input: "/bookstore/book[position() < 3]",
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
				NewToken("3", TokenInt),
				NewToken("]", TokenRBracket),
			},
		},
		{
			name:  "node with a binary and-expression in the attribute",
			input: "/bookstore/book[@price > 35.00 && @price < 50.00]",
			want: []*Token{
				NewToken("/", TokenSlash),
				NewToken("bookstore", TokenNode),
				NewToken("/", TokenSlash),
				NewToken("book", TokenNode),
				NewToken("[", TokenLBracket),
				NewToken("@", TokenAt),
				NewToken("price", TokenNode),
				NewToken(">", TokenGreater),
				NewToken("35.00", TokenFloat),
				NewToken("&&", TokenAnd),
				NewToken("@", TokenAt),
				NewToken("price", TokenNode),
				NewToken("<", TokenLess),
				NewToken("50.00", TokenFloat),
				NewToken("]", TokenRBracket),
			},
		},
		{
			name:  "node with a binary or-expression in the attribute",
			input: "/bookstore/book[@price > 35.00 || @price < 50.00]",
			want: []*Token{
				NewToken("/", TokenSlash),
				NewToken("bookstore", TokenNode),
				NewToken("/", TokenSlash),
				NewToken("book", TokenNode),
				NewToken("[", TokenLBracket),
				NewToken("@", TokenAt),
				NewToken("price", TokenNode),
				NewToken(">", TokenGreater),
				NewToken("35.00", TokenFloat),
				NewToken("||", TokenOr),
				NewToken("@", TokenAt),
				NewToken("price", TokenNode),
				NewToken("<", TokenLess),
				NewToken("50.00", TokenFloat),
				NewToken("]", TokenRBracket),
			},
		},
		{
			name:  "node with a basic function call",
			input: "/books[length()-1]",
			want: []*Token{
				NewToken("/", TokenSlash),
				NewToken("books", TokenNode),
				NewToken("[", TokenLBracket),
				NewToken("length", TokenNode),
				NewToken("(", TokenLParen),
				NewToken(")", TokenRParen),
				NewToken("-", TokenMinus),
				NewToken("1", TokenInt),
				NewToken("]", TokenRBracket),
			},
		},
		{
			name:  "node with a binary operator",
			input: "/books[2 + 3]",
			want: []*Token{
				NewToken("/", TokenSlash),
				NewToken("books", TokenNode),
				NewToken("[", TokenLBracket),
				NewToken("2", TokenInt),
				NewToken("+", TokenPlus),
				NewToken("3", TokenInt),
				NewToken("]", TokenRBracket),
			},
		},
		{
			name:  "node with a function call in a bool expr",
			input: "/books[position() <= 1]",
			want: []*Token{
				NewToken("/", TokenSlash),
				NewToken("books", TokenNode),
				NewToken("[", TokenLBracket),
				NewToken("position", TokenNode),
				NewToken("(", TokenLParen),
				NewToken(")", TokenRParen),
				NewToken("<=", TokenLessEqual),
				NewToken("1", TokenInt),
				NewToken("]", TokenRBracket),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := tokenizeXPathQuery(tt.input)
			if !errorsSimilar(tt.wantErr, err) {
				t.Errorf("parseXPathQuery() error = %v, want %v", err, tt.wantErr)
				return
			}
			if tt.wantErr != nil {
				return
			}

			if !deepEqual(tt.want, tokens) {
				t.Errorf("Unexpected result: want: %+v, got: %+v", tt.want, tokens)
			}
		})
	}
}
