package protoquery

import (
	"fmt"
	reflect "reflect"
	"testing"
)

func TestcompileQuery(t *testing.T) {
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
			wantErr: fmt.Errorf("unexpected prefix token ]"),
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
		{
			name: "node with key and attribute filter",
			input: []*Token{
				NewToken("nodename", TokenNode),
				NewToken("[", TokenLBracket),
				NewToken("key", TokenString),
				NewToken("]", TokenRBracket),
				NewToken("/", TokenSlash),
				NewToken("prop", TokenNode),
			},
			want: Query{
				&NodeQueryStep{
					name: "nodename",
				},
				&KeyQueryStep{
					expr: &LiteralExpr{
						value: "key",
						typ:   TypeString,
					},
				},
				&NodeQueryStep{
					name: "prop",
				},
			},
		},
		{
			name: "path with a root recursive descent operator",
			input: []*Token{
				NewToken("//", TokenSlashSlash),
				NewToken("node", TokenNode),
			},
			want: Query{
				&RecursiveDescentQueryStep{},
				&NodeQueryStep{
					name: "node",
				},
			},
		},
		{
			name: "path with a middle-sitter recursive descent operator",
			input: []*Token{
				NewToken("node", TokenNode),
				NewToken("//", TokenSlashSlash),
				NewToken("ancestor", TokenNode),
			},
			want: Query{
				&NodeQueryStep{
					name: "node",
				},
				&RecursiveDescentQueryStep{},
				&NodeQueryStep{
					name: "ancestor",
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
