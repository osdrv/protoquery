package protoquery

import "fmt"

type (
	parsePrefixFn func([]*Token, int, int) (Expression, int, error)
	parseInfixFn  func([]*Token, int, Expression, int) (Expression, int, error)
)

const (
	LOWEST int = iota
	EQUALS
	COMPARE
	SUM
	MULTIPLY
	PREFIX
)

var (
	parsePrefixFns = map[TokenKind]parsePrefixFn{}
	parseInfixFns  = map[TokenKind]parseInfixFn{}

	TokenOpMap = map[TokenKind]Operator{
		TokenBang:         OpNot,
		TokenPlus:         OpPlus,
		TokenMinus:        OpMinus,
		TokenStar:         OpMul,
		TokenSlash:        OpDiv,
		TokenEqual:        OpEq,
		TokenNotEqual:     OpNe,
		TokenAnd:          OpAnd,
		TokenOr:           OpOr,
		TokenLess:         OpLt,
		TokenLessEqual:    OpLe,
		TokenGreater:      OpGt,
		TokenGreaterEqual: OpGe,
	}

	precedences = map[TokenKind]int{
		TokenEqual:        EQUALS,
		TokenNotEqual:     EQUALS,
		TokenLess:         COMPARE,
		TokenLessEqual:    COMPARE,
		TokenGreater:      COMPARE,
		TokenGreaterEqual: COMPARE,
		TokenPlus:         SUM,
		TokenMinus:        SUM,
		TokenAnd:          SUM,
		TokenOr:           SUM,
		TokenSlash:        MULTIPLY,
		TokenStar:         MULTIPLY,
	}
)

func init() {
	parsePrefixFns[TokenAt] = parsePropertyExpression
	parsePrefixFns[TokenNode] = parseFunctionExpression

	parsePrefixFns[TokenInt] = parseLiteralExpression
	parsePrefixFns[TokenFloat] = parseLiteralExpression
	parsePrefixFns[TokenBool] = parseLiteralExpression
	parsePrefixFns[TokenString] = parseLiteralExpression

	parsePrefixFns[TokenBang] = parseUnaryExpression
	parsePrefixFns[TokenPlus] = parseUnaryExpression
	parsePrefixFns[TokenMinus] = parseUnaryExpression
	parsePrefixFns[TokenLParen] = parseGroupExpression

	parseInfixFns[TokenEqual] = parseBinaryExpression
	parseInfixFns[TokenNotEqual] = parseBinaryExpression
	parseInfixFns[TokenLess] = parseBinaryExpression
	parseInfixFns[TokenLessEqual] = parseBinaryExpression
	parseInfixFns[TokenGreater] = parseBinaryExpression
	parseInfixFns[TokenGreaterEqual] = parseBinaryExpression
	parseInfixFns[TokenPlus] = parseBinaryExpression
	parseInfixFns[TokenMinus] = parseBinaryExpression
	parseInfixFns[TokenSlash] = parseBinaryExpression
	parseInfixFns[TokenStar] = parseBinaryExpression
	parseInfixFns[TokenAnd] = parseBinaryExpression
	parseInfixFns[TokenOr] = parseBinaryExpression
}

func parseExpression(tokens []*Token, ix int, precedence int) (Expression, int, error) {
	prefix, ok := parsePrefixFns[tokens[ix].Kind]
	if !ok {
		return nil, ix, fmt.Errorf("unexpected prefix token %v", tokens[ix].Value)
	}
	var leftExpr Expression
	var err error
	leftExpr, ix, err = prefix(tokens, ix, precedence)

	if ix < len(tokens) && precedence < precedences[tokens[ix].Kind] {
		infix, ok := parseInfixFns[tokens[ix].Kind]
		if !ok {
			return nil, ix, fmt.Errorf("unexpected infix token %v", tokens[ix].Value)
		}
		leftExpr, ix, err = infix(tokens, ix, leftExpr, precedence)
	}

	return leftExpr, ix, err
}

func parsePropertyExpression(tokens []*Token, ix int, precedence int) (Expression, int, error) {
	ix++
	if !matchTokenAny(tokens, ix, TokenNode, TokenStar) {
		return nil, ix, fmt.Errorf("expected node or '*', got %v", tokens[ix].Value)
	}
	return &PropertyExpr{tokens[ix].Value}, ix + 1, nil
}

func parseFunctionExpression(tokens []*Token, ix int, precedence int) (Expression, int, error) {
	expr := &FunctionCallExpr{
		handle: tokens[ix].Value,
	}
	ix++
	if !matchToken(tokens, ix, TokenLParen) {
		return nil, ix, fmt.Errorf("expected '(', got %v", tokens[ix].Value)
	}
	ix++
	if !matchToken(tokens, ix, TokenRParen) {
		args := []Expression{}
		for ix < len(tokens) && !matchToken(tokens, ix, TokenRParen) {
			var arg Expression
			var err error
			arg, ix, err = parseExpression(tokens, ix, LOWEST)
			if err != nil {
				return nil, ix, err
			}
			args = append(args, arg)
			if !matchToken(tokens, ix, TokenComma) {
				break
			}
			ix++
		}
		expr.args = args
	}
	if !matchToken(tokens, ix, TokenRParen) {
		return nil, ix, fmt.Errorf("expected ')', got %v", tokens[ix].Value)
	}
	ix++
	return expr, ix, nil
}

func parseLiteralExpression(tokens []*Token, ix int, precedence int) (Expression, int, error) {
	expr := &LiteralExpr{}
	switch tokens[ix].Kind {
	case TokenInt:
		intv, err := tokens[ix].IntValue()
		if err != nil {
			return nil, ix, err
		}
		expr.value = intv
		expr.typ = TypeInt
	case TokenFloat:
		floatv, err := tokens[ix].FloatValue()
		if err != nil {
			return nil, ix, err
		}
		expr.value = floatv
		expr.typ = TypeFloat
	case TokenBool:
		boolv, err := tokens[ix].BoolValue()
		if err != nil {
			return nil, ix, err
		}
		expr.value = boolv
		expr.typ = TypeBool
	default:
		expr.value = tokens[ix].Value
		expr.typ = TypeString
	}
	return expr, ix + 1, nil
}

func parseUnaryExpression(tokens []*Token, ix int, precedence int) (Expression, int, error) {
	op, ok := TokenOpMap[tokens[ix].Kind]
	if !ok {
		return nil, ix, fmt.Errorf("unexpected prefix token %v", tokens[ix].Value)
	}
	var rightExpr Expression
	var err error
	rightExpr, ix, err = parseExpression(tokens, ix+1, PREFIX)
	if err != nil {
		return nil, ix, err
	}
	expr := &UnaryExpr{
		op:   op,
		expr: rightExpr,
	}
	return expr, ix, nil
}

func parseGroupExpression(tokens []*Token, ix int, precedence int) (Expression, int, error) {
	var group Expression
	var err error
	group, ix, err = parseExpression(tokens, ix+1, LOWEST)
	if err != nil {
		return nil, ix, err
	}
	if tokens[ix].Kind != TokenRParen {
		return nil, ix, fmt.Errorf("expected ')', got %v", tokens[ix].Value)
	}
	return group, ix + 1, nil
}

func parseBinaryExpression(tokens []*Token, ix int, left Expression, precedence int) (Expression, int, error) {
	op, ok := TokenOpMap[tokens[ix].Kind]
	if !ok {
		return nil, ix, fmt.Errorf("undefined operator for infix token %v", tokens[ix].Value)
	}
	expr := &BinaryExpr{
		left: left,
		op:   op,
	}
	var right Expression
	var err error
	right, ix, err = parseExpression(tokens, ix+1, precedence)
	if err != nil {
		return nil, ix, err
	}
	expr.right = right

	return expr, ix, nil
}
