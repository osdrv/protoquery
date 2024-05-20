package protoquery

import (
	"strings"
	"testing"
)

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

func TestFindAll(t *testing.T) {
	store := Bookstore{
		Books: []*Book{
			{
				Title:  "The Go Programming Language",
				Author: "Alan A. A. Donovan",
				Price:  34.99,
			},
			{
				Title:  "The Rust Programming Language",
				Author: "Steve Klabnik",
				Price:  39.99,
			},
		},
	}

	pq, err := Compile("/bookstore/book[price>35.00]")
	if err != nil {
		t.Fatalf("Compile() error = %v, no error expected", err)
	}

	res := pq.FindAll(&store)
	if len(res) != 1 {
		t.Errorf("FindAll() got = %v, want 1", len(res))
	}
}
