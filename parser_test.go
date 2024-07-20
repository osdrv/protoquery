package protoquery

import (
	"errors"
	"testing"
)

func TestParseExpression(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []*Token
		want    Expression
		wantErr error
	}{
		{
			name: "numeric literal",
			tokens: []*Token{
				NewToken("123", TokenInt),
			},
			want: &LiteralExpr{
				value: int64(123),
				typ:   TypeInt,
			},
		},
		{
			name: "string literal",
			tokens: []*Token{
				NewToken("value", TokenString),
			},
			want: &LiteralExpr{
				value: "value",
				typ:   TypeString,
			},
		},
		{
			name: "boolean literal",
			tokens: []*Token{
				NewToken("true", TokenBool),
			},
			want: &LiteralExpr{
				value: true,
				typ:   TypeBool,
			},
		},
		{
			name: "named property",
			tokens: []*Token{
				NewToken("@", TokenAt),
				NewToken("prop", TokenNode),
			},
			want: &PropertyExpr{
				name: "prop",
			},
		},
		{
			name: "wildcard property",
			tokens: []*Token{
				NewToken("@", TokenAt),
				NewToken("*", TokenStar),
			},
			want: &PropertyExpr{
				name: "*",
			},
		},
		{
			name: "basic function call",
			tokens: []*Token{
				NewToken("func", TokenNode),
				NewToken("(", TokenLParen),
				NewToken(")", TokenRParen),
			},
			want: &FunctionCallExpr{
				handle: "func",
			},
		},
		{
			name: "function call with arguments",
			tokens: []*Token{
				NewToken("func", TokenNode),
				NewToken("(", TokenLParen),
				NewToken("42", TokenInt),
				NewToken(",", TokenComma),
				NewToken("@", TokenAt),
				NewToken("prop", TokenNode),
				NewToken(",", TokenComma),
				NewToken("true", TokenBool),
				NewToken(",", TokenComma),
				NewToken("another_func", TokenNode),
				NewToken("(", TokenLParen),
				NewToken(")", TokenRParen),
				NewToken(")", TokenRParen),
			},
			want: &FunctionCallExpr{
				handle: "func",
				args: []Expression{
					&LiteralExpr{
						value: int64(42),
						typ:   TypeInt,
					},
					&PropertyExpr{
						name: "prop",
					},
					&LiteralExpr{
						value: true,
						typ:   TypeBool,
					},
					&FunctionCallExpr{
						handle: "another_func",
					},
				},
			},
		},
		{
			name: "bool unary expr",
			tokens: []*Token{
				NewToken("!", TokenBang),
				NewToken("@", TokenAt),
				NewToken("bool_prop", TokenNode),
			},
			want: &UnaryExpr{
				expr: &PropertyExpr{
					name: "bool_prop",
				},
				op: OpNot,
			},
		},
		{
			name: "unary plus expr",
			tokens: []*Token{
				NewToken("+", TokenPlus),
				NewToken("42", TokenInt),
			},
			want: &UnaryExpr{
				expr: &LiteralExpr{
					value: int64(42),
					typ:   TypeInt,
				},
				op: OpPlus,
			},
		},
		{
			name: "unary minus expr",
			tokens: []*Token{
				NewToken("-", TokenMinus),
				NewToken("42", TokenInt),
			},
			want: &UnaryExpr{
				expr: &LiteralExpr{
					value: int64(42),
					typ:   TypeInt,
				},
				op: OpMinus,
			},
		},
		{
			name: "binary expr with 2 literals",
			tokens: []*Token{
				NewToken("42", TokenInt),
				NewToken("+", TokenPlus),
				NewToken("24", TokenInt),
			},
			want: &BinaryExpr{
				left: &LiteralExpr{
					value: int64(42),
					typ:   TypeInt,
				},
				right: &LiteralExpr{
					value: int64(24),
					typ:   TypeInt,
				},
				op: OpPlus,
			},
		},
		{
			name: "binary with operator precedence",
			tokens: []*Token{
				NewToken("42", TokenInt),
				NewToken("+", TokenPlus),
				NewToken("24", TokenInt),
				NewToken("*", TokenStar),
				NewToken("2", TokenInt),
			},
			want: &BinaryExpr{
				left: &LiteralExpr{
					value: int64(42),
					typ:   TypeInt,
				},
				right: &BinaryExpr{
					left: &LiteralExpr{
						value: int64(24),
						typ:   TypeInt,
					},
					right: &LiteralExpr{
						value: int64(2),
						typ:   TypeInt,
					},
					op: OpMul,
				},
				op: OpPlus,
			},
		},
		{
			name: "binary with parentheses",
			tokens: []*Token{
				NewToken("(", TokenLParen),
				NewToken("42", TokenInt),
				NewToken("+", TokenPlus),
				NewToken("24", TokenInt),
				NewToken(")", TokenRParen),
				NewToken("*", TokenStar),
				NewToken("2", TokenInt),
			},
			want: &BinaryExpr{
				left: &BinaryExpr{
					left: &LiteralExpr{
						value: int64(42),
						typ:   TypeInt,
					},
					right: &LiteralExpr{
						value: int64(24),
						typ:   TypeInt,
					},
					op: OpPlus,
				},
				right: &LiteralExpr{
					value: int64(2),
					typ:   TypeInt,
				},
				op: OpMul,
			},
		},
		{
			name: "comparison expression with precedence",
			tokens: []*Token{
				NewToken("42", TokenInt),
				NewToken("<=", TokenLessEqual),
				NewToken("24", TokenInt),
				NewToken("+", TokenPlus),
				NewToken("2", TokenInt),
			},
			want: &BinaryExpr{
				left: &LiteralExpr{
					value: int64(42),
					typ:   TypeInt,
				},
				right: &BinaryExpr{
					left: &LiteralExpr{
						value: int64(24),
						typ:   TypeInt,
					},
					right: &LiteralExpr{
						value: int64(2),
						typ:   TypeInt,
					},
					op: OpPlus,
				},
				op: OpLe,
			},
		},
		{
			name: "binary unary with parenteses",
			tokens: []*Token{
				NewToken("!", TokenBang),
				NewToken("(", TokenLParen),
				NewToken("true", TokenBool),
				NewToken("&&", TokenAnd),
				NewToken("false", TokenBool),
				NewToken(")", TokenRParen),
			},
			want: &UnaryExpr{
				expr: &BinaryExpr{
					left: &LiteralExpr{
						value: true,
						typ:   TypeBool,
					},
					right: &LiteralExpr{
						value: false,
						typ:   TypeBool,
					},
					op: OpAnd,
				},
				op: OpNot,
			},
		},
		{
			name: "unary operation with property",
			tokens: []*Token{
				NewToken("-", TokenMinus),
				NewToken("@", TokenAt),
				NewToken("prop", TokenNode),
			},
			want: &UnaryExpr{
				expr: &PropertyExpr{
					name: "prop",
				},
				op: OpMinus,
			},
		},
		{
			name: "binary and unary expressions with precedence",
			tokens: []*Token{
				NewToken("!", TokenBang),
				NewToken("true", TokenBool),
				NewToken("&&", TokenAnd),
				NewToken("false", TokenBool),
			},
			want: &BinaryExpr{
				left: &UnaryExpr{
					expr: &LiteralExpr{
						value: true,
						typ:   TypeBool,
					},
					op: OpNot,
				},
				right: &LiteralExpr{
					value: false,
					typ:   TypeBool,
				},
				op: OpAnd,
			},
		},
		{
			name: "binary and unary expressions with precedence with parentheses",
			tokens: []*Token{
				NewToken("!", TokenBang),
				NewToken("(", TokenLParen),
				NewToken("true", TokenBool),
				NewToken("&&", TokenAnd),
				NewToken("false", TokenBool),
				NewToken(")", TokenRParen),
			},
			want: &UnaryExpr{
				expr: &BinaryExpr{
					left: &LiteralExpr{
						value: true,
						typ:   TypeBool,
					},
					right: &LiteralExpr{
						value: false,
						typ:   TypeBool,
					},
					op: OpAnd,
				},
				op: OpNot,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := parseExpression(tt.tokens, 0, LOWEST)
			if !errorsSimilar(err, tt.wantErr) {
				t.Errorf("parseExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr != nil {
				return
			}
			if !deepEqual(got, tt.want) {
				t.Errorf("parseExpression() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseLiteralExpression(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []*Token
		want    Expression
		wantErr error
	}{
		{
			name: "int",
			tokens: []*Token{
				{Kind: TokenInt, Value: "42"},
			},
			want: &LiteralExpr{
				value: 42,
				typ:   TypeInt,
			},
		},
		{
			name: "float",
			tokens: []*Token{
				{Kind: TokenFloat, Value: "42.0"},
			},
			want: &LiteralExpr{
				value: 42.0,
				typ:   TypeFloat,
			},
		},
		{
			name: "bool",
			tokens: []*Token{
				{Kind: TokenBool, Value: "true"},
			},
			want: &LiteralExpr{
				value: true,
				typ:   TypeBool,
			},
		},
		{
			name: "string",
			tokens: []*Token{
				{Kind: TokenString, Value: "hello"},
			},
			want: &LiteralExpr{
				value: "hello",
				typ:   TypeString,
			},
		},
		{name: "error", tokens: []*Token{
			{Kind: TokenComma, Value: ","},
		},
			wantErr: errors.New("unexpected prefix token"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := parseExpression(tt.tokens, 0, LOWEST)
			if !errorsSimilar(err, tt.wantErr) {
				t.Errorf("parseExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr != nil {
				return
			}
			if !deepEqual(got, tt.want) {
				t.Errorf("parseExpression() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsePrefixExpression(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []*Token
		want    Expression
		wantErr error
	}{
		{
			name: "bang",
			tokens: []*Token{
				{Kind: TokenBang, Value: "!"},
				{Kind: TokenBool, Value: "true"},
			},
			want: &UnaryExpr{
				op: OpNot,
				expr: &LiteralExpr{
					value: true,
					typ:   TypeBool,
				},
			},
		},
		{
			name: "plus",
			tokens: []*Token{
				{Kind: TokenPlus, Value: "+"},
				{Kind: TokenInt, Value: "42"},
			},
			want: &UnaryExpr{
				op: OpPlus,
				expr: &LiteralExpr{
					value: 42,
					typ:   TypeInt,
				},
			},
		},
		{
			name: "minus",
			tokens: []*Token{
				{Kind: TokenMinus, Value: "-"},
				{Kind: TokenInt, Value: "42"},
			},
			want: &UnaryExpr{
				op: OpMinus,
				expr: &LiteralExpr{
					value: 42,
					typ:   TypeInt,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := parseExpression(tt.tokens, 0, LOWEST)
			if !errorsSimilar(err, tt.wantErr) {
				t.Errorf("parseExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr != nil {
				return
			}
			if !deepEqual(got, tt.want) {
				t.Errorf("parseExpression() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseBinaryExpression(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []*Token
		want    Expression
		wantErr error
	}{
		{
			name: "add",
			tokens: []*Token{
				{Kind: TokenInt, Value: "1"},
				{Kind: TokenPlus, Value: "+"},
				{Kind: TokenInt, Value: "2"},
			},
			want: &BinaryExpr{
				op:    OpPlus,
				left:  &LiteralExpr{value: 1, typ: TypeInt},
				right: &LiteralExpr{value: 2, typ: TypeInt},
			},
		},
		{
			name: "minus",
			tokens: []*Token{
				{Kind: TokenInt, Value: "1"},
				{Kind: TokenMinus, Value: "-"},
				{Kind: TokenInt, Value: "2"},
			},
			want: &BinaryExpr{
				op:    OpMinus,
				left:  &LiteralExpr{value: 1, typ: TypeInt},
				right: &LiteralExpr{value: 2, typ: TypeInt},
			},
		},
		{
			name: "multiply",
			tokens: []*Token{
				{Kind: TokenInt, Value: "1"},
				{Kind: TokenStar, Value: "*"},
				{Kind: TokenInt, Value: "2"},
			},
			want: &BinaryExpr{
				op:    OpMul,
				left:  &LiteralExpr{value: 1, typ: TypeInt},
				right: &LiteralExpr{value: 2, typ: TypeInt},
			},
		},
		{
			name: "divide",
			tokens: []*Token{
				{Kind: TokenInt, Value: "1"},
				{Kind: TokenSlash, Value: "/"},
				{Kind: TokenInt, Value: "2"},
			},
			want: &BinaryExpr{
				op:    OpDiv,
				left:  &LiteralExpr{value: 1, typ: TypeInt},
				right: &LiteralExpr{value: 2, typ: TypeInt},
			},
		},
		{
			name: "plus multiply",
			tokens: []*Token{
				{Kind: TokenInt, Value: "1"},
				{Kind: TokenPlus, Value: "+"},
				{Kind: TokenInt, Value: "2"},
				{Kind: TokenStar, Value: "*"},
				{Kind: TokenInt, Value: "3"},
			},
			want: &BinaryExpr{
				op:   OpPlus,
				left: &LiteralExpr{value: 1, typ: TypeInt},
				right: &BinaryExpr{
					op:    OpMul,
					left:  &LiteralExpr{value: 2, typ: TypeInt},
					right: &LiteralExpr{value: 3, typ: TypeInt},
				},
			},
		},
		{
			name: "plus with prefix",
			tokens: []*Token{
				{Kind: TokenMinus, Value: "-"},
				{Kind: TokenInt, Value: "1"},
				{Kind: TokenPlus, Value: "+"},
				{Kind: TokenMinus, Value: "-"},
				{Kind: TokenInt, Value: "2"},
			},
			want: &BinaryExpr{
				op:    OpPlus,
				left:  &UnaryExpr{op: OpMinus, expr: &LiteralExpr{value: 1, typ: TypeInt}},
				right: &UnaryExpr{op: OpMinus, expr: &LiteralExpr{value: 2, typ: TypeInt}},
			},
		},
		{
			name: "binary op with prefix",
			tokens: []*Token{
				{Kind: TokenBang, Value: "!"},
				{Kind: TokenBool, Value: "true"},
				{Kind: TokenOr, Value: "||"},
				{Kind: TokenBang, Value: "!"},
				{Kind: TokenBool, Value: "false"},
			},
			want: &BinaryExpr{
				op:    OpOr,
				left:  &UnaryExpr{op: OpNot, expr: &LiteralExpr{value: true, typ: TypeBool}},
				right: &UnaryExpr{op: OpNot, expr: &LiteralExpr{value: false, typ: TypeBool}},
			},
		},
		{
			name: "arithmetic and boolean",
			tokens: []*Token{
				{Kind: TokenInt, Value: "1"},
				{Kind: TokenPlus, Value: "+"},
				{Kind: TokenInt, Value: "2"},
				{Kind: TokenLess, Value: "<"},
				{Kind: TokenInt, Value: "3"},
			},
			want: &BinaryExpr{
				op: OpLt,
				left: &BinaryExpr{
					op:    OpPlus,
					left:  &LiteralExpr{value: 1, typ: TypeInt},
					right: &LiteralExpr{value: 2, typ: TypeInt},
				},
				right: &LiteralExpr{value: 3, typ: TypeInt},
			},
		},
		{
			name: "grouping with precedence",
			tokens: []*Token{
				{Kind: TokenLParen, Value: "("},
				{Kind: TokenInt, Value: "1"},
				{Kind: TokenPlus, Value: "+"},
				{Kind: TokenInt, Value: "2"},
				{Kind: TokenRParen, Value: ")"},
				{Kind: TokenStar, Value: "*"},
				{Kind: TokenInt, Value: "3"},
			},
			want: &BinaryExpr{
				op: OpMul,
				left: &BinaryExpr{
					op:    OpPlus,
					left:  &LiteralExpr{value: 1, typ: TypeInt},
					right: &LiteralExpr{value: 2, typ: TypeInt},
				},
				right: &LiteralExpr{value: 3, typ: TypeInt},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := parseExpression(tt.tokens, 0, LOWEST)
			if !errorsSimilar(err, tt.wantErr) {
				t.Errorf("parseExpression() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr != nil {
				return
			}
			if !deepEqual(got, tt.want) {
				t.Errorf("parseExpression() = %v, want %v", got, tt.want)
			}
		})
	}
}
