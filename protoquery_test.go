package protoquery

import (
	"reflect"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
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

	tests := []struct {
		name    string
		query   string
		want    []proto.Message
		wantErr error
	}{
		{
			name:  "children elements",
			query: "/books/book",
			want: []proto.Message{
				store.Books[0],
				store.Books[1],
			},
		},
		{
			name:  "child element by numeric index",
			query: "/books/book[1]",
			want: []proto.Message{
				store.Books[1],
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pq, err := Compile(tt.query)
			if !errorsSimilar(tt.wantErr, err) {
				t.Errorf("Compile() error = %v, want %v", err, tt.wantErr)
				return
			}
			if err != nil {
				t.Errorf("Compile() error = %v, no error expected", err)
				return
			}
			res := pq.FindAll(&store)
			if !reflect.DeepEqual(res, tt.want) {
				t.Errorf("Compile() = %+v, want %+v", res, tt.want)
			}
		})
	}

	//pq, err := Compile("/books/book[@price>35.00]")
	//if err != nil {
	//	t.Fatalf("Compile() error = %v, no error expected", err)
	//}

	//res := pq.FindAll(&store)
	//if len(res) != 1 {
	//	t.Errorf("FindAll() got = %v, want 1", len(res))
	//}
}
