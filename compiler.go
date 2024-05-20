package protoquery

import "fmt"

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

var (
	tokenToCmp = map[TokenKind]AttrCmp{
		TokenEqual:        AttrCmpEq,
		TokenNotEqual:     AttrCmpNe,
		TokenLess:         AttrCmpLt,
		TokenLessEqual:    AttrCmpLe,
		TokenGreater:      AttrCmpGt,
		TokenGreaterEqual: AttrCmpGe,
	}
)

func compilePredicate(tokens []*Token, ix int) (Predicate, int, error) {
	if !matchToken(tokens, ix, TokenLBracket) {
		return nil, ix, fmt.Errorf("expected [, got %v", tokens[ix].Kind)
	}
	ix++
	var p Predicate
	if matchToken(tokens, ix, TokenNumber) {
		index, err := tokens[ix].IntValue()
		if err != nil {
			return nil, 0, err
		}
		p = &IndexPredicate{
			Index: index,
		}
		ix++
	} else {
		ap := &AttrPredicate{}
		if !matchToken(tokens, ix, TokenAt) {
			return nil, ix, fmt.Errorf("expected @, got %v", tokens[ix].Kind)
		}
		ix++
		if !matchToken(tokens, ix, TokenNode) {
			return nil, ix, fmt.Errorf("expected attribute name, got %v", tokens[ix].Kind)
		}
		ap.Name = tokens[ix].Value
		ap.Cmp = AttrCmpExist
		ix++

		if matchTokenAny(tokens, ix, TokenEqual, TokenGreater, TokenGreaterEqual,
			TokenNotEqual, TokenLess, TokenLessEqual) {
			ap.Cmp = tokenToCmp[tokens[ix].Kind]
			ix++
			if !matchTokenAny(tokens, ix, TokenString, TokenNumber) {
				return nil, ix, fmt.Errorf("expected string or number, got %v", tokens[ix].Kind)
			}
			// TODO(osdrv): are we loosing the type information here?
			ap.Value = tokens[ix].Value
			ix++
		}
		p = ap
	}
	if !matchToken(tokens, ix, TokenRBracket) {
		return nil, ix, fmt.Errorf("expected ], got %v", tokens[ix].Kind)
	}
	ix++
	return p, ix, nil
}

func compileNodeQueryStep(tokens []*Token, ix int) (*NodeQueryStep, int, error) {
	nqs := &NodeQueryStep{}
	if !matchToken(tokens, ix, TokenNode) {
		return nil, ix, fmt.Errorf("expected node name, got %v", tokens[ix].Kind)
	}
	nqs.name = tokens[ix].Value
	ix++
	var err error
	for ix < len(tokens) && matchToken(tokens, ix, TokenLBracket) {
		var p Predicate
		p, ix, err = compilePredicate(tokens, ix)
		if err != nil {
			return nil, ix, err
		}
		if nqs.predicate == nil {
			nqs.predicate = p
		} else {
			nqs.predicate = &AndPredicate{
				predicates: []Predicate{nqs.predicate, p},
			}
		}
	}
	return nqs, ix, nil
}

func compileQuery(tokens []*Token) (Query, error) {
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
			var nqs *NodeQueryStep
			nqs, ix, err = compileNodeQueryStep(tokens, ix)
			if err != nil {
				return nil, err
			}
			query = append(query, nqs)
		default:
			return nil, fmt.Errorf("unexpected token %v %q", tokens[ix].Kind, tokens[ix].Value)
		}
	}
	return query, nil
}
