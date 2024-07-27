package protoquery

import (
	"errors"
	"testing"

	"github.com/osdrv/protoquery/proto"
)

func TestExpressionType(t *testing.T) {
	msg := &proto.Book{
		Title:  "The Go Programming Language",
		Author: "Alan A. A. Donovan",
		Price:  42.99,
		Pages:  432,
		OnSale: true,
	}
	tests := []struct {
		name    string
		input   Expression
		ctx     EvalContext
		want    Type
		wantErr error
	}{
		{
			name:  "literal with string type",
			input: NewLiteralExpr("hello", TypeString),
			ctx:   NewEvalContext(nil),
			want:  TypeString,
		},
		{
			name:  "literal with integer type",
			input: NewLiteralExpr(42, TypeInt),
			ctx:   NewEvalContext(nil),
			want:  TypeInt,
		},
		{
			name:  "literal with float type",
			input: NewLiteralExpr(42.99, TypeFloat),
			ctx:   NewEvalContext(nil),
			want:  TypeFloat,
		},
		{
			name:  "literal with bool type",
			input: NewLiteralExpr(true, TypeBool),
			ctx:   NewEvalContext(nil),
			want:  TypeBool,
		},
		{
			name: "property with string type",
			input: &PropertyExpr{
				name: "author",
			},
			ctx:  NewEvalContext(msg.ProtoReflect()),
			want: TypeString,
		},
		{
			name: "property with number type",
			input: &PropertyExpr{
				name: "pages",
			},
			ctx:  NewEvalContext(msg.ProtoReflect()),
			want: TypeInt,
		},
		{
			name: "property with bool type",
			input: &PropertyExpr{
				name: "on_sale",
			},
			ctx:  NewEvalContext(msg.ProtoReflect()),
			want: TypeBool,
		},
		{
			name: "property with float type",
			input: &PropertyExpr{
				name: "price",
			},
			ctx:  NewEvalContext(msg.ProtoReflect()),
			want: TypeFloat,
		},
		{
			name: "property with EnforceBool=true",
			input: &PropertyExpr{
				name: "price",
			},
			ctx:  NewEvalContext(msg.ProtoReflect(), WithEnforceBool(true)),
			want: TypeBool,
		},
		{
			name: "binary addition expression with integers",
			input: &BinaryExpr{
				left:  NewLiteralExpr(42, TypeInt),
				right: NewLiteralExpr(2, TypeInt),
				op:    OpPlus,
			},
			ctx:  NewEvalContext(nil),
			want: TypeInt,
		},
		{
			name: "binary comparison with integers",
			input: &BinaryExpr{
				left:  NewLiteralExpr(42, TypeInt),
				right: NewLiteralExpr(2, TypeInt),
				op:    OpGt,
			},
			ctx:  NewEvalContext(nil),
			want: TypeBool,
		},
		{
			name: "binary comparison with floats",
			input: &BinaryExpr{
				left:  NewLiteralExpr(42.99, TypeFloat),
				right: NewLiteralExpr(2.0, TypeFloat),
				op:    OpGt,
			},
			ctx:  NewEvalContext(nil),
			want: TypeBool,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.Type(tt.ctx)
			if !errorsSimilar(err, tt.wantErr) {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("unexpected type: got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExpressionEval(t *testing.T) {
	msg := &proto.Book{
		Title:  "The Go Programming Language",
		Author: "Alan A. A. Donovan",
		Price:  42.99,
	}

	msgEmpty := &proto.Book{}

	tests := []struct {
		name    string
		input   Expression
		ctx     EvalContext
		want    any
		wantErr error
	}{
		{
			name:  "literal integer",
			input: NewLiteralExpr(42, TypeInt),
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
			name:  "literal float",
			input: NewLiteralExpr(42.99, TypeFloat),
			ctx:   NewEvalContext(nil),
			want:  float64(42.99),
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
			name: "existing property with EnforceBool=true",
			input: &PropertyExpr{
				name: "price",
			},
			ctx:  NewEvalContext(msg.ProtoReflect(), WithEnforceBool(true)),
			want: true,
		},
		{
			name: "existing unset property with EnforceBool=true",
			input: &PropertyExpr{
				name: "price",
			},
			ctx:  NewEvalContext(msgEmpty.ProtoReflect(), WithEnforceBool(true)),
			want: false,
		},
		{
			name: "non existing property with EnforceBool=true",
			input: &PropertyExpr{
				name: "non_existing",
			},
			ctx:     NewEvalContext(msg.ProtoReflect(), WithEnforceBool(true)),
			wantErr: PropNotSet,
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
			name: "unset string property with UseDefault=false",
			input: &PropertyExpr{
				name: "author",
			},
			ctx:     NewEvalContext(msgEmpty.ProtoReflect(), WithUseDefault(false)),
			wantErr: PropNotSet,
		},
		{
			name: "unset string property with UseDefault=true",
			input: &PropertyExpr{
				name: "author",
			},
			ctx:  NewEvalContext(msgEmpty.ProtoReflect(), WithUseDefault(true)),
			want: "",
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
				expr: NewLiteralExpr(42, TypeInt),
				op:   OpMinus,
			},
			ctx:  NewEvalContext(nil),
			want: int64(-42),
		},
		{
			name: "unary numeric expression +42",
			input: &UnaryExpr{
				expr: NewLiteralExpr(42, TypeInt),
				op:   OpPlus,
			},
			ctx:  NewEvalContext(nil),
			want: int64(42),
		},
		{
			name: "binary numeric expression +",
			input: &BinaryExpr{
				left:  NewLiteralExpr(40, TypeInt),
				right: NewLiteralExpr(2, TypeInt),
				op:    OpPlus,
			},
			ctx:  NewEvalContext(nil),
			want: int64(42),
		},
		{
			name: "binary numeric expression *",
			input: &BinaryExpr{
				left:  NewLiteralExpr(40, TypeInt),
				right: NewLiteralExpr(2, TypeInt),
				op:    OpMul,
			},
			ctx:  NewEvalContext(nil),
			want: int64(80),
		},
		{
			name: "binary numeric expression -",
			input: &BinaryExpr{
				left:  NewLiteralExpr(40, TypeInt),
				right: NewLiteralExpr(2, TypeInt),
				op:    OpMinus,
			},
			ctx:  NewEvalContext(nil),
			want: int64(38),
		},
		{
			name: "binary numeric expression /",
			input: &BinaryExpr{
				left:  NewLiteralExpr(40, TypeInt),
				right: NewLiteralExpr(2, TypeInt),
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
		{
			name: "binary bool expression =",
			input: &BinaryExpr{
				left:  NewLiteralExpr(true, TypeBool),
				right: NewLiteralExpr(false, TypeBool),
				op:    OpEq,
			},
			ctx:  NewEvalContext(nil),
			want: false,
		},
		{
			name: "binary bool expression !=",
			input: &BinaryExpr{
				left:  NewLiteralExpr(true, TypeBool),
				right: NewLiteralExpr(false, TypeBool),
				op:    OpNe,
			},
			ctx:  NewEvalContext(nil),
			want: true,
		},
		{
			name: "binary bool expression !=",
			input: &BinaryExpr{
				left:  NewLiteralExpr(true, TypeBool),
				right: NewLiteralExpr(false, TypeBool),
				op:    OpNe,
			},
			ctx:  NewEvalContext(nil),
			want: true,
		},
		{
			name: "binary numeric expression =",
			input: &BinaryExpr{
				left:  NewLiteralExpr(1, TypeInt),
				right: NewLiteralExpr(2, TypeInt),
				op:    OpEq,
			},
			ctx:  NewEvalContext(nil),
			want: false,
		},
		{
			name: "binary numeric expression !=",
			input: &BinaryExpr{
				left:  NewLiteralExpr(1, TypeInt),
				right: NewLiteralExpr(2, TypeInt),
				op:    OpNe,
			},
			ctx:  NewEvalContext(nil),
			want: true,
		},
		{
			name: "binary numeric expression >",
			input: &BinaryExpr{
				left:  NewLiteralExpr(1, TypeInt),
				right: NewLiteralExpr(2, TypeInt),
				op:    OpGt,
			},
			ctx:  NewEvalContext(nil),
			want: false,
		},
		{
			name: "binary numeric expression >=",
			input: &BinaryExpr{
				left:  NewLiteralExpr(1, TypeInt),
				right: NewLiteralExpr(2, TypeInt),
				op:    OpGe,
			},
			ctx:  NewEvalContext(nil),
			want: false,
		},
		{
			name: "binary numeric expression <",
			input: &BinaryExpr{
				left:  NewLiteralExpr(1, TypeInt),
				right: NewLiteralExpr(2, TypeInt),
				op:    OpLt,
			},
			ctx:  NewEvalContext(nil),
			want: true,
		},
		{
			name: "binary numeric expression <=",
			input: &BinaryExpr{
				left:  NewLiteralExpr(1, TypeInt),
				right: NewLiteralExpr(2, TypeInt),
				op:    OpLt,
			},
			ctx:  NewEvalContext(nil),
			want: true,
		},
		{
			name: "binary string expression unequal =",
			input: &BinaryExpr{
				left:  NewLiteralExpr("foo", TypeString),
				right: NewLiteralExpr("bar", TypeString),
				op:    OpEq,
			},
			ctx:  NewEvalContext(nil),
			want: false,
		},
		{
			name: "binary string expression unequal !=",
			input: &BinaryExpr{
				left:  NewLiteralExpr("foo", TypeString),
				right: NewLiteralExpr("bar", TypeString),
				op:    OpNe,
			},
			ctx:  NewEvalContext(nil),
			want: true,
		},
		{
			name: "binary string expression equal =",
			input: &BinaryExpr{
				left:  NewLiteralExpr("foo", TypeString),
				right: NewLiteralExpr("foo", TypeString),
				op:    OpEq,
			},
			ctx:  NewEvalContext(nil),
			want: true,
		},
		{
			name: "binary string expression >",
			input: &BinaryExpr{
				left:  NewLiteralExpr("foo", TypeString),
				right: NewLiteralExpr("bar", TypeString),
				op:    OpGt,
			},
			ctx:  NewEvalContext(nil),
			want: true,
		},
		{
			name: "binary string expression >=",
			input: &BinaryExpr{
				left:  NewLiteralExpr("foo", TypeString),
				right: NewLiteralExpr("bar", TypeString),
				op:    OpGe,
			},
			ctx:  NewEvalContext(nil),
			want: true,
		},
		{
			name: "binary string expression <",
			input: &BinaryExpr{
				left:  NewLiteralExpr("foo", TypeString),
				right: NewLiteralExpr("bar", TypeString),
				op:    OpLt,
			},
			ctx:  NewEvalContext(nil),
			want: false,
		},
		{
			name: "binary string expression <=",
			input: &BinaryExpr{
				left:  NewLiteralExpr("foo", TypeString),
				right: NewLiteralExpr("bar", TypeString),
				op:    OpLe,
			},
			ctx:  NewEvalContext(nil),
			want: false,
		},
		{
			name: "binary string expression -",
			input: &BinaryExpr{
				left:  NewLiteralExpr("foo", TypeString),
				right: NewLiteralExpr("bar", TypeString),
				op:    OpMinus,
			},
			ctx:     NewEvalContext(nil),
			wantErr: errors.New("Invalid type"),
		},
		{
			name: "binary string expression *",
			input: &BinaryExpr{
				left:  NewLiteralExpr("foo", TypeString),
				right: NewLiteralExpr("bar", TypeString),
				op:    OpMul,
			},
			ctx:     NewEvalContext(nil),
			wantErr: errors.New("Invalid type"),
		},
		{
			name: "binary string expression /",
			input: &BinaryExpr{
				left:  NewLiteralExpr("foo", TypeString),
				right: NewLiteralExpr("bar", TypeString),
				op:    OpDiv,
			},
			ctx:     NewEvalContext(nil),
			wantErr: errors.New("Invalid type"),
		},
		{
			name: "binary expression with floats",
			input: &BinaryExpr{
				left:  NewLiteralExpr(1.0, TypeFloat),
				right: NewLiteralExpr(2.0, TypeFloat),
				op:    OpPlus,
			},
			ctx:  NewEvalContext(nil),
			want: 3.0,
		},
		{
			name: "binary comparison expression with floats",
			input: &BinaryExpr{
				left:  NewLiteralExpr(1.0, TypeFloat),
				right: NewLiteralExpr(2.0, TypeFloat),
				op:    OpLt,
			},
			ctx:  NewEvalContext(nil),
			want: true,
		},
		{
			name: "binary expression = with empty context object and EnforceBool=true",
			input: &BinaryExpr{
				left:  NewLiteralExpr(2, TypeInt),
				right: NewLiteralExpr(2, TypeInt),
				op:    OpEq,
			},
			ctx:  NewEvalContext(nil, WithEnforceBool(true)),
			want: true,
		},
		{
			name: "binary expression = with non-empty context object and EnforceBool=true",
			input: &BinaryExpr{
				left:  NewPropertyExpr("title"),
				right: NewLiteralExpr("The Go Programming Language", TypeString),
				op:    OpEq,
			},
			ctx:  NewEvalContext(msg.ProtoReflect(), WithEnforceBool(true)),
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

func TestIsAllPropertyExprs(t *testing.T) {
	tests := []struct {
		name  string
		input Expression
		want  bool
	}{
		{
			name: "PropertyExpr",
			input: &PropertyExpr{
				name: "foo",
			},
			want: true,
		},
		{
			name: "LiteralExpr",
			input: &LiteralExpr{
				value: int64(1),
				typ:   TypeInt,
			},
			want: false,
		},
		{
			name: "BinaryExpr with PropertyExprs",
			input: &BinaryExpr{
				left: &PropertyExpr{
					name: "foo",
				},
				right: &PropertyExpr{
					name: "bar",
				},
				op: OpAnd,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAllPropertyExprs(tt.input); got != tt.want {
				t.Errorf("IsAllPropertyExprs() = %v, want %v", got, tt.want)
			}
		})
	}
}
