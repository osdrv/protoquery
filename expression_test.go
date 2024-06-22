package protoquery

import (
	"testing"
)

func TestExpressionEval(t *testing.T) {
	msg := &Book{
		Title:  "The Go Programming Language",
		Author: "Alan A. A. Donovan",
		Price:  42.99,
	}

	tests := []struct {
		name    string
		input   Expression
		ctx     *EvalContext
		want    any
		wantErr error
	}{
		{
			name:  "literal number",
			input: NewLiteralExpr(42, TypeNumber),
			ctx:   NewEvalContext(nil),
			want:  int64(42),
		},
		{
			name:  "literal string",
			input: NewLiteralExpr("hello", TypeString),
			ctx:   NewEvalContext(nil),
			want:  "hello",
		},
		{
			name:  "literal bool",
			input: NewLiteralExpr(true, TypeBool),
			ctx:   NewEvalContext(nil),
			want:  true,
		},
		{
			name: "property string",
			input: &PropertyExpr{
				name: "author",
			},
			ctx:  NewEvalContext(msg.ProtoReflect()),
			want: "Alan A. A. Donovan",
		},
		{
			name: "property number",
			input: &PropertyExpr{
				name: "price",
			},
			ctx:  NewEvalContext(msg.ProtoReflect()),
			want: float32(42.99),
		},
		{
			name: "non existing property",
			input: &PropertyExpr{
				name: "non_existing",
			},
			ctx:     NewEvalContext(msg.ProtoReflect()),
			wantErr: PropNotSet,
		},
		{
			name: "unary boolean expression !true",
			input: &UnaryExpr{
				expr: NewLiteralExpr(true, TypeBool),
				op:   OpNot,
			},
			ctx:  NewEvalContext(nil),
			want: false,
		},
		{
			name: "unary boolean expression !false",
			input: &UnaryExpr{
				expr: NewLiteralExpr(false, TypeBool),
				op:   OpNot,
			},
			ctx:  NewEvalContext(nil),
			want: true,
		},
		{
			name: "unary numeric expression -42",
			input: &UnaryExpr{
				expr: NewLiteralExpr(42, TypeNumber),
				op:   OpMinus,
			},
			ctx:  NewEvalContext(nil),
			want: int64(-42),
		},
		{
			name: "unary numeric expression +42",
			input: &UnaryExpr{
				expr: NewLiteralExpr(42, TypeNumber),
				op:   OpPlus,
			},
			ctx:  NewEvalContext(nil),
			want: int64(42),
		},
		{
			name: "binary numeric expression +",
			input: &BinaryExpr{
				left:  NewLiteralExpr(40, TypeNumber),
				right: NewLiteralExpr(2, TypeNumber),
				op:    OpPlus,
			},
			ctx:  NewEvalContext(nil),
			want: int64(42),
		},
		{
			name: "binary numeric expression *",
			input: &BinaryExpr{
				left:  NewLiteralExpr(40, TypeNumber),
				right: NewLiteralExpr(2, TypeNumber),
				op:    OpMul,
			},
			ctx:  NewEvalContext(nil),
			want: int64(80),
		},
		{
			name: "binary numeric expression -",
			input: &BinaryExpr{
				left:  NewLiteralExpr(40, TypeNumber),
				right: NewLiteralExpr(2, TypeNumber),
				op:    OpMinus,
			},
			ctx:  NewEvalContext(nil),
			want: int64(38),
		},
		{
			name: "binary numeric expression /",
			input: &BinaryExpr{
				left:  NewLiteralExpr(40, TypeNumber),
				right: NewLiteralExpr(2, TypeNumber),
				op:    OpDiv,
			},
			ctx:  NewEvalContext(nil),
			want: int64(20),
		},
		{
			name: "binary string expression +",
			input: &BinaryExpr{
				left:  NewLiteralExpr("foo", TypeString),
				right: NewLiteralExpr("bar", TypeString),
				op:    OpPlus,
			},
			ctx:  NewEvalContext(nil),
			want: "foobar",
		},
		{
			name: "binary bool expression &&",
			input: &BinaryExpr{
				left:  NewLiteralExpr(true, TypeBool),
				right: NewLiteralExpr(false, TypeBool),
				op:    OpAnd,
			},
			ctx:  NewEvalContext(nil),
			want: false,
		},
		{
			name: "binary bool expression ||",
			input: &BinaryExpr{
				left:  NewLiteralExpr(true, TypeBool),
				right: NewLiteralExpr(false, TypeBool),
				op:    OpOr,
			},
			ctx:  NewEvalContext(nil),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.Eval(tt.ctx)
			if !errorsSimilar(err, tt.wantErr) {
				t.Errorf("LiteralExpr.Eval() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if !deepEqual(got, tt.want) {
				t.Errorf("LiteralExpr.Eval() = %v, want %v", got, tt.want)
			}
		})
	}
}
