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
			want: &Literal{
				value: int64(123),
				typ:   TypeNumber,
			},
		},
		{
			name: "string literal",
			input: []*Token{
				NewToken("value", TokenString),
			},
			want: &Literal{
				value: "value",
				typ:   TypeString,
			},
		},
		{
			name: "boolean literal",
			input: []*Token{
				NewToken("true", TokenBool),
			},
			want: &Literal{
				value: true,
				typ:   TypeBool,
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
