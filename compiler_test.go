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
				&KeyQueryStep{
					expr: &PropertyExpr{
						name: "attr",
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
				&KeyQueryStep{
					expr: &BinaryExpr{
						left: &PropertyExpr{
							name: "attr",
						},
						right: &LiteralExpr{
							value: "value",
							typ:   TypeString,
						},
						op: OpEq,
					},
				},
			},
		},
		{
			name: "node with attribute comparison",
			input: []*Token{
				NewToken("nodename", TokenNode),
				NewToken("[", TokenLBracket),
				NewToken("@", TokenAt),
				NewToken("attr", TokenNode),
				NewToken(">", TokenGreater),
				NewToken("35", TokenInt),
				NewToken("]", TokenRBracket),
			},
			want: Query{
				&NodeQueryStep{
					name: "nodename",
				},
				&KeyQueryStep{
					expr: &BinaryExpr{
						left: &PropertyExpr{
							name: "attr",
						},
						right: &LiteralExpr{
							value: int64(35),
							typ:   TypeInt,
						},
						op: OpGt,
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
			wantErr: fmt.Errorf("unexpected token ]"),
		},
		{
			name: "node with an index dereference and an attribute filter",
			input: []*Token{
				NewToken("nodename", TokenNode),
				NewToken("[", TokenLBracket),
				NewToken("1", TokenInt),
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
					expr: &LiteralExpr{
						value: int64(1),
						typ:   TypeInt,
					},
				},
				&KeyQueryStep{
					expr: &BinaryExpr{
						left: &PropertyExpr{
							name: "attr",
						},
						right: &LiteralExpr{
							value: "value",
							typ:   TypeString,
						},
						op: OpEq,
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
		{
			name: "node with an attribute with a function call",
			input: []*Token{
				NewToken("nodename", TokenNode),
				NewToken("[", TokenLBracket),
				NewToken("position", TokenNode),
				NewToken("(", TokenLParen),
				NewToken(")", TokenRParen),
				NewToken("<=", TokenLessEqual),
				NewToken("10", TokenInt),
				NewToken("]", TokenRBracket),
			},
			want: Query{
				&NodeQueryStep{
					name: "nodename",
				},
				&KeyQueryStep{
					expr: &BinaryExpr{
						left: &FunctionCallExpr{
							handle: "position",
						},
						right: &LiteralExpr{
							value: int64(10),
							typ:   TypeInt,
						},
						op: OpLe,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CompileQuery(tt.input)
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
				NewToken("123", TokenInt),
			},
			want: &LiteralExpr{
				value: int64(123),
				typ:   TypeInt,
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
			input: []*Token{
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
			input: []*Token{
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
			input: []*Token{
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
			input: []*Token{
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
			input: []*Token{
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
			got, _, err := CompileExpression(tt.input, 0)
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
