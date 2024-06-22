package protoquery

import (
	"fmt"
	reflect "reflect"
	"testing"
)

func TestCompileQuery(t *testing.T) {
	tests := []struct {
		name    string
		input   []*Token
		want    Query
		wantErr error
	}{
		{
			name: "plain node query",
			input: []*Token{
				NewToken("nodename", TokenNode),
			},
			want: Query{
				&NodeQueryStep{name: "nodename"},
			},
		},
		{
			name: "node with attribute presence",
			input: []*Token{
				NewToken("nodename", TokenNode),
				NewToken("[", TokenLBracket),
				NewToken("@", TokenAt),
				NewToken("attr", TokenNode),
				NewToken("]", TokenRBracket),
			},
			want: Query{
				&NodeQueryStep{
					name: "nodename",
				},
				&AttrFilterQueryStep{
					predicate: &AttrPredicate{
						Name: "attr",
						Cmp:  AttrCmpExist,
					},
				},
			},
		},
		{
			name: "node with attribute equality",
			input: []*Token{
				NewToken("nodename", TokenNode),
				NewToken("[", TokenLBracket),
				NewToken("@", TokenAt),
				NewToken("attr", TokenNode),
				NewToken("=", TokenEqual),
				NewToken("value", TokenString),
				NewToken("]", TokenRBracket),
			},
			want: Query{
				&NodeQueryStep{
					name: "nodename",
				},
				&AttrFilterQueryStep{
					predicate: &AttrPredicate{
						Name:  "attr",
						Cmp:   AttrCmpEq,
						Value: "value",
					},
				},
			},
		},
		{
			name: "node with attribute equality but a missing operand",
			input: []*Token{
				NewToken("nodename", TokenNode),
				NewToken("[", TokenLBracket),
				NewToken("@", TokenAt),
				NewToken("attr", TokenNode),
				NewToken("=", TokenEqual),
				NewToken("]", TokenRBracket),
			},
			wantErr: fmt.Errorf("expected string or number"),
		},
		{
			name: "node with an index dereference and an attribute filter",
			input: []*Token{
				NewToken("nodename", TokenNode),
				NewToken("[", TokenLBracket),
				NewToken("1", TokenNumber),
				NewToken("]", TokenRBracket),
				NewToken("[", TokenLBracket),
				NewToken("@", TokenAt),
				NewToken("attr", TokenNode),
				NewToken("=", TokenEqual),
				NewToken("value", TokenString),
				NewToken("]", TokenRBracket),
			},
			want: Query{
				&NodeQueryStep{
					name: "nodename",
				},
				&KeyQueryStep{
					Term:  "1",
					IsNum: true,
					Num:   1,
				},
				&AttrFilterQueryStep{
					predicate: &AttrPredicate{
						Name:  "attr",
						Cmp:   AttrCmpEq,
						Value: "value",
					},
				},
			},
		},
		{
			name: "node with a full path from root",
			input: []*Token{
				NewToken("/", TokenSlash),
				NewToken("parent", TokenNode),
				NewToken("/", TokenSlash),
				NewToken("child", TokenNode),
			},
			want: Query{
				&RootQueryStep{},
				&NodeQueryStep{
					name: "parent",
				},
				&NodeQueryStep{
					name: "child",
				},
			},
		},
		{
			name: "node with a recursive descent search",
			input: []*Token{
				NewToken("/", TokenSlash),
				NewToken("parent", TokenNode),
				NewToken("//", TokenSlashSlash),
				NewToken("child", TokenNode),
			},
			want: Query{
				&RootQueryStep{},
				&NodeQueryStep{
					name: "parent",
				},
				&RecursiveDescentQueryStep{},
				&NodeQueryStep{
					name: "child",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := compileQuery(tt.input)
			if tt.wantErr != nil {
				if !errorsSimilar(err, tt.wantErr) {
					t.Errorf("compileQuery() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("compileQuery() error = %v, no error expected", err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("compileQuery() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestCompileExpression(t *testing.T) {
	tests := []struct {
		name    string
		input   []*Token
		want    Expression
		wantErr error
	}{
		{
			name: "numeric literal",
			input: []*Token{
				NewToken("123", TokenNumber),
			},
			want: &LiteralExpr{
				value: int64(123),
				typ:   TypeNumber,
			},
		},
		{
			name: "string literal",
			input: []*Token{
				NewToken("value", TokenString),
			},
			want: &LiteralExpr{
				value: "value",
				typ:   TypeString,
			},
		},
		{
			name: "boolean literal",
			input: []*Token{
				NewToken("true", TokenBool),
			},
			want: &LiteralExpr{
				value: true,
				typ:   TypeBool,
			},
		},
		{
			name: "named property",
			input: []*Token{
				NewToken("@", TokenAt),
				NewToken("prop", TokenNode),
			},
			want: &PropertyExpr{
				name: "prop",
			},
		},
		{
			name: "wildcard property",
			input: []*Token{
				NewToken("@", TokenAt),
				NewToken("*", TokenStar),
			},
			want: &PropertyExpr{
				name: "*",
			},
		},
		{
			name: "basic function call",
			input: []*Token{
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
			input: []*Token{
				NewToken("func", TokenNode),
				NewToken("(", TokenLParen),
				NewToken("42", TokenNumber),
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
						typ:   TypeNumber,
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
			input: []*Token{
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
			input: []*Token{
				NewToken("+", TokenPlus),
				NewToken("42", TokenNumber),
			},
			want: &UnaryExpr{
				expr: &LiteralExpr{
					value: int64(42),
					typ:   TypeNumber,
				},
				op: OpPlus,
			},
		},
		{
			name: "unary minus expr",
			input: []*Token{
				NewToken("-", TokenMinus),
				NewToken("42", TokenNumber),
			},
			want: &UnaryExpr{
				expr: &LiteralExpr{
					value: int64(42),
					typ:   TypeNumber,
				},
				op: OpMinus,
			},
		},
		{
			name: "binary expr with 2 literals",
			input: []*Token{
				NewToken("42", TokenNumber),
				NewToken("+", TokenPlus),
				NewToken("24", TokenNumber),
			},
			want: &BinaryExpr{
				left: &LiteralExpr{
					value: int64(42),
					typ:   TypeNumber,
				},
				right: &LiteralExpr{
					value: int64(24),
					typ:   TypeNumber,
				},
				op: OpPlus,
			},
		},
		{
			name: "binary with operator precedence",
			input: []*Token{
				NewToken("42", TokenNumber),
				NewToken("+", TokenPlus),
				NewToken("24", TokenNumber),
				NewToken("*", TokenStar),
				NewToken("2", TokenNumber),
			},
			want: &BinaryExpr{
				left: &LiteralExpr{
					value: int64(42),
					typ:   TypeNumber,
				},
				right: &BinaryExpr{
					left: &LiteralExpr{
						value: int64(24),
						typ:   TypeNumber,
					},
					right: &LiteralExpr{
						value: int64(2),
						typ:   TypeNumber,
					},
					op: OpMul,
				},
				op: OpPlus,
			},
		},
		{
			name: "binary with parentheses",
			input: []*Token{
				NewToken("(", TokenLParen),
				NewToken("42", TokenNumber),
				NewToken("+", TokenPlus),
				NewToken("24", TokenNumber),
				NewToken(")", TokenRParen),
				NewToken("*", TokenStar),
				NewToken("2", TokenNumber),
			},
			want: &BinaryExpr{
				left: &BinaryExpr{
					left: &LiteralExpr{
						value: int64(42),
						typ:   TypeNumber,
					},
					right: &LiteralExpr{
						value: int64(24),
						typ:   TypeNumber,
					},
					op: OpPlus,
				},
				right: &LiteralExpr{
					value: int64(2),
					typ:   TypeNumber,
				},
				op: OpMul,
			},
		},
		{
			name: "comparison expression with precedence",
			input: []*Token{
				NewToken("42", TokenNumber),
				NewToken("<=", TokenLessEqual),
				NewToken("24", TokenNumber),
				NewToken("+", TokenPlus),
				NewToken("2", TokenNumber),
			},
			want: &BinaryExpr{
				left: &LiteralExpr{
					value: int64(42),
					typ:   TypeNumber,
				},
				right: &BinaryExpr{
					left: &LiteralExpr{
						value: int64(24),
						typ:   TypeNumber,
					},
					right: &LiteralExpr{
						value: int64(2),
						typ:   TypeNumber,
					},
					op: OpPlus,
				},
				op: OpLe,
			},
		},
		{
			name: "binary unary with parenteses",
			input: []*Token{
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
			input: []*Token{
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
			input: []*Token{
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
			input: []*Token{
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
			got, err := CompileExpression(tt.input)
			if tt.wantErr != nil {
				if !errorsSimilar(err, tt.wantErr) {
					t.Errorf("compileExpression() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("compileExpression() error = %v, no error expected", err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("compileExpression() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
