package protoquery

import "strings"

// errorEqual compares two errors. It returns true if both are nil,
// or if both are not nil and have the same error message.
func errorEqual(err1, err2 error) bool {
	if err1 == nil && err2 == nil {
		return true
	}
	if err1 == nil || err2 == nil {
		return false
	}
	return err1.Error() == err2.Error()
}

// errorSimilar compares two errors. It returns true if both are nil,
// or if both are not nil and the error message of err1 contains the error message of err2.
func errorsSimilar(err1, err2 error) bool {
	if err1 == nil || err2 == nil {
		return err1 == err2
	}
	return strings.Contains(err1.Error(), err2.Error())
}
