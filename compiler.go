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

func compileIndexQueryStep(tokens []*Token, ix int) (*IndexQueryStep, int, error) {
	iqs := &IndexQueryStep{}
	if !matchToken(tokens, ix, TokenNumber) {
		return nil, ix, fmt.Errorf("expected number, got %v", tokens[ix].Kind)
	}
	index, err := tokens[ix].IntValue()
	if err != nil {
		return nil, 0, err
	}
	iqs.index = index
	ix++
	return iqs, ix, nil
}

func compileAttrFilterStep(tokens []*Token, ix int) (*AttrFilterStep, int, error) {
	afs := &AttrFilterStep{}
	if !matchToken(tokens, ix, TokenAt) {
		return nil, ix, fmt.Errorf("expected @, got %v", tokens[ix].Kind)
	}
	ix++
	if !matchToken(tokens, ix, TokenNode) {
		return nil, ix, fmt.Errorf("expected attribute name, got %v", tokens[ix].Kind)
	}
	afs.predicate = &AttrPredicate{
		Name: tokens[ix].Value,
		Cmp:  AttrCmpExist,
	}
	ix++

	if matchTokenAny(tokens, ix, TokenEqual, TokenGreater, TokenGreaterEqual,
		TokenNotEqual, TokenLess, TokenLessEqual) {
		afs.predicate.Cmp = tokenToCmp[tokens[ix].Kind]
		ix++
		if !matchTokenAny(tokens, ix, TokenString, TokenNumber) {
			return nil, ix, fmt.Errorf("expected string or number, got %v", tokens[ix].Kind)
		}
		// TODO(osdrv): are we loosing the type information here?
		afs.predicate.Value = tokens[ix].Value
		ix++
	}
	return afs, ix, nil
}

func compileNodeQueryStep(tokens []*Token, ix int) (*NodeQueryStep, int, error) {
	nqs := &NodeQueryStep{}
	if !matchToken(tokens, ix, TokenNode) {
		return nil, ix, fmt.Errorf("expected node name, got %v", tokens[ix].Kind)
	}
	nqs.name = tokens[ix].Value
	ix++
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
		case TokenLBracket:
			ix++
			if matchToken(tokens, ix, TokenNumber) {
				// Index query step
				var iqs QueryStep
				iqs, ix, err = compileIndexQueryStep(tokens, ix)
				if err != nil {
					return nil, err
				}
				query = append(query, iqs)
			} else if matchToken(tokens, ix, TokenAt) {
				// Attribute filter step
				// Keep it duplicated becase at some point I want to
				// unwrap AND-filter in a series of separate filters.
				var afs QueryStep
				afs, ix, err = compileAttrFilterStep(tokens, ix)
				if err != nil {
					return nil, err
				}
				query = append(query, afs)
			} else {
				return nil, fmt.Errorf("unexpected token %v", tokens[ix].Kind)
			}
			if !matchToken(tokens, ix, TokenRBracket) {
				return nil, fmt.Errorf("expected ], got %v", tokens[ix].Kind)
			}
			ix++
		default:
			return nil, fmt.Errorf("unexpected token %v %q", tokens[ix].Kind, tokens[ix].Value)
		}
	}
	return query, nil
}
