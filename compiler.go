package protoquery

import (
	"fmt"
)

var (
	tokenToOp = map[TokenKind]Operator{
		TokenPlus:         OpPlus,
		TokenMinus:        OpMinus,
		TokenAnd:          OpAnd,
		TokenOr:           OpOr,
		TokenStar:         OpMul,
		TokenSlash:        OpDiv,
		TokenEqual:        OpEq,
		TokenNotEqual:     OpNe,
		TokenLess:         OpLt,
		TokenLessEqual:    OpLe,
		TokenGreater:      OpGt,
		TokenGreaterEqual: OpGe,
	}
)

func matchToken(tokens []*Token, ix int, kind TokenKind) bool {
	return ix < len(tokens) && tokens[ix].Kind == kind
}

func matchTokenAny(tokens []*Token, ix int, kinds ...TokenKind) bool {
	for _, kind := range kinds {
		if matchToken(tokens, ix, kind) {
			return true
		}
	}
	return false
}

func compileNodeQueryStep(tokens []*Token, ix int) (*NodeQueryStep, int, error) {
	nqs := &NodeQueryStep{}
	if !matchToken(tokens, ix, TokenNode) {
		return nil, ix, fmt.Errorf("expected node name, got %v", tokens[ix].Value)
	}
	nqs.name = tokens[ix].Value
	ix++
	return nqs, ix, nil
}

func compileKeyQueryStep(tokens []*Token, ix int) (*KeyQueryStep, int, error) {
	var expr Expression
	var err error
	if !matchToken(tokens, ix, TokenLBracket) {
		return nil, ix, fmt.Errorf("expected [, got %v", tokens[ix].Value)
	}
	ix++
	expr, ix, err = CompileExpression(tokens, ix)
	if err != nil {
		return nil, ix, err
	}
	if !matchToken(tokens, ix, TokenRBracket) {
		return nil, ix, fmt.Errorf("expected ], got %v", tokens[ix].Value)
	}
	ix++
	return &KeyQueryStep{
		expr: expr,
	}, ix, nil
}

func CompileQuery(tokens []*Token) (Query, error) {
	var query Query
	ix := 0
	for ix < len(tokens) {
		var err error
		switch tokens[ix].Kind {
		case TokenSlash:
			if len(query) > 0 {
				// Separator between nodes
				ix++
				continue
			}
			ix++
			query = append(query, &RootQueryStep{})
		case TokenSlashSlash:
			query = append(query, &RecursiveDescentQueryStep{})
			ix++
		case TokenNode:
			var qs *NodeQueryStep
			qs, ix, err = compileNodeQueryStep(tokens, ix)
			if err != nil {
				return nil, err
			}
			query = append(query, qs)
		case TokenLBracket:
			var qs QueryStep
			qs, ix, err = compileKeyQueryStep(tokens, ix)
			if err != nil {
				return nil, err
			}
			query = append(query, qs)
		default:
			return nil, fmt.Errorf("unexpected token %v %q", tokens[ix].Kind, tokens[ix].Value)
		}
	}
	return query, nil
}

func CompileExpression(tokens []*Token, ix int) (Expression, int, error) {
	return compileComparisonExpression(tokens, ix)
}

func compileComparisonExpression(tokens []*Token, ix int) (Expression, int, error) {
	var err error
	var expr Expression
	expr, ix, err = compileAdditionExpression(tokens, ix)
	if err != nil {
		return nil, ix, err
	}
	for matchTokenAny(tokens, ix, TokenEqual, TokenNotEqual, TokenLess, TokenLessEqual, TokenGreater, TokenGreaterEqual) {
		op, ok := tokenToOp[tokens[ix].Kind]
		if !ok {
			return nil, ix, fmt.Errorf("unexpected comparison token %v", tokens[ix].Value)
		}
		ix++
		var right Expression
		right, ix, err = compileAdditionExpression(tokens, ix)
		if err != nil {
			return nil, ix, err
		}
		expr = &BinaryExpr{
			left:  expr,
			right: right,
			op:    op,
		}
	}
	return expr, ix, nil
}

func compileAdditionExpression(tokens []*Token, ix int) (Expression, int, error) {
	var err error
	var expr Expression
	expr, ix, err = compileMultiplyExpression(tokens, ix)
	if err != nil {
		return nil, ix, err
	}
	for matchTokenAny(tokens, ix, TokenPlus, TokenMinus, TokenAnd, TokenOr) {
		op, ok := tokenToOp[tokens[ix].Kind]
		if !ok {
			return nil, ix, fmt.Errorf("unexpected addition token %v", tokens[ix].Value)
		}
		ix++
		var right Expression
		right, ix, err = compileMultiplyExpression(tokens, ix)
		if err != nil {
			return nil, ix, err
		}
		expr = &BinaryExpr{
			left:  expr,
			right: right,
			op:    op,
		}
	}
	return expr, ix, nil
}

func compileMultiplyExpression(tokens []*Token, ix int) (Expression, int, error) {
	var err error
	var expr Expression
	expr, ix, err = compileElementaryExpression(tokens, ix)
	if err != nil {
		return nil, ix, err
	}
	for matchTokenAny(tokens, ix, TokenStar, TokenSlash) {
		op, ok := tokenToOp[tokens[ix].Kind]
		if !ok {
			return nil, ix, fmt.Errorf("unexpected multiply token %v", tokens[ix].Value)
		}
		ix++
		var right Expression
		right, ix, err = compileElementaryExpression(tokens, ix)
		if err != nil {
			return nil, ix, err
		}
		expr = &BinaryExpr{
			left:  expr,
			right: right,
			op:    op,
		}
	}
	return expr, ix, nil
}

func compileElementaryExpression(tokens []*Token, ix int) (Expression, int, error) {
	var expr Expression
	var err error
	switch tokens[ix].Kind {
	case TokenAt:
		expr, ix, err = compilePropertyExpression(tokens, ix)
	case TokenNode:
		expr, ix, err = compileFunctionCallExpression(tokens, ix)
	case TokenBang, TokenPlus, TokenMinus:
		expr, ix, err = compileUnaryExpression(tokens, ix)
	case TokenInt, TokenFloat, TokenBool, TokenString:
		expr, ix, err = compileLiteralExpression(tokens, ix)
	case TokenLParen:
		ix++
		expr, ix, err = compileComparisonExpression(tokens, ix)
		if err != nil {
			return nil, ix, err
		}
		if !matchToken(tokens, ix, TokenRParen) {
			return nil, ix, fmt.Errorf("expected ')', got %v", tokens[ix].Value)
		}
		ix++
	default:
		return nil, ix, fmt.Errorf("unexpected token %v", tokens[ix].Value)
	}

	return expr, ix, err
}

func compileLiteralExpression(tokens []*Token, ix int) (Expression, int, error) {
	if !matchTokenAny(tokens, ix, TokenString, TokenInt, TokenFloat, TokenBool) {
		return nil, ix, fmt.Errorf("expected string or number, got %v", tokens[ix].Value)
	}
	var val any
	val = tokens[ix].Value
	typ := TypeString
	switch tokens[ix].Kind {
	case TokenString:
	case TokenInt:
		intv, err := tokens[ix].IntValue()
		if err != nil {
			return nil, ix, err
		}
		val = intv
		typ = TypeInt
	case TokenFloat:
		floatv, err := tokens[ix].FloatValue()
		if err != nil {
			return nil, ix, err
		}
		val = floatv
		typ = TypeFloat
	case TokenBool:
		boolv, err := tokens[ix].BoolValue()
		if err != nil {
			return nil, ix, err
		}
		val = boolv
		typ = TypeBool
	default:
		return nil, ix, fmt.Errorf("unexpected token %v %q", tokens[ix].Kind, tokens[ix].Value)
	}
	return &LiteralExpr{
		value: val,
		typ:   typ,
	}, ix + 1, nil
}

func compilePropertyExpression(tokens []*Token, ix int) (Expression, int, error) {
	if !matchToken(tokens, ix, TokenAt) {
		return nil, ix, fmt.Errorf("expected @, got %v", tokens[ix].Value)
	}
	ix++
	if !matchTokenAny(tokens, ix, TokenNode, TokenStar) {
		return nil, ix, fmt.Errorf("expected attribute name, got %v", tokens[ix].Value)
	}
	return &PropertyExpr{
		name: tokens[ix].Value,
	}, ix + 1, nil
}

func compileFunctionCallExpression(tokens []*Token, ix int) (Expression, int, error) {
	if !matchToken(tokens, ix, TokenNode) {
		return nil, ix, fmt.Errorf("expected function name, got %v", tokens[ix].Value)
	}
	name := tokens[ix].Value
	ix++
	if !matchToken(tokens, ix, TokenLParen) {
		return nil, ix, fmt.Errorf("expected (, got %v", tokens[ix].Value)
	}
	ix++
	var args []Expression
	for ix < len(tokens) {
		if matchToken(tokens, ix, TokenRParen) {
			break
		}
		var arg Expression
		var err error
		arg, ix, err = compileComparisonExpression(tokens, ix)
		if err != nil {
			return nil, ix, err
		}
		args = append(args, arg)
		if !matchToken(tokens, ix, TokenComma) {
			break
		}
		ix++
	}
	if !matchToken(tokens, ix, TokenRParen) {
		return nil, ix, fmt.Errorf("expected ), got %v", tokens[ix].Value)
	}
	return &FunctionCallExpr{
		handle: name,
		args:   args,
	}, ix + 1, nil
}

func compileUnaryExpression(tokens []*Token, ix int) (Expression, int, error) {
	var op Operator
	switch tokens[ix].Kind {
	case TokenBang:
		op = OpNot
	case TokenPlus:
		op = OpPlus
	case TokenMinus:
		op = OpMinus
	default:
		return nil, ix, fmt.Errorf("unexpected token %v %q", tokens[ix].Kind, tokens[ix].Value)
	}
	ix++
	var expr Expression
	var err error
	expr, ix, err = compileElementaryExpression(tokens, ix)
	if err != nil {
		return nil, ix, err
	}
	return &UnaryExpr{
		op:   op,
		expr: expr,
	}, ix, nil
}
