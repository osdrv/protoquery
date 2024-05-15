package protoquery

import "testing"

// errorEqual compares two errors. It returns true if both are nil, or if both are not nil and have the same error message.
func errorEqual(err1, err2 error) bool {
	if err1 == nil && err2 == nil {
		return true
	}
	if err1 == nil || err2 == nil {
		return false
	}
	return err1.Error() == err2.Error()
}

// func TestCompile(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		input   string
// 		wantErr error
// 	}{}
//
// 	for _, tt := range tests {
// 		t.Run("Compile", func(t *testing.T) {
// 			got, err := Compile()
// 			if tt.wantErr != nil {
// 				if !errorEqual(err, tt.wantErr) {
// 					t.Fatalf("Compile() error = %v, want %v", err, tt.wantErr)
// 				}
// 				return
// 			}
// 			if err != nil {
// 				t.Fatalf("Compile() error = %v, want nil", err)
// 			}
// 			if got == nil {
// 				t.Fatalf("Compile() got = nil, want not nil")
// 			}
// 		})
// 	}
// }

func TestFindAll(t *testing.T) {
	ab := AddressBook{
		People: []*Person{
			{
				Id:    1,
				Name:  "Alice",
				Email: "alice@evilcorp.com",
				Phones: []*Person_PhoneNumber{
					{
						Type:   PhoneType_PHONE_TYPE_HOME,
						Number: "+1234567890",
					},
					{
						Type:   PhoneType_PHONE_TYPE_MOBILE,
						Number: "+1234567890",
					},
				},
			},
		},
	}

	pq, err := Compile("people[0]/phones[@type='mobile']")
	if err != nil {
		t.Fatalf("Compile() error = %v, want nil", err)
	}

	res := pq.FindAll(&ab)
	if res == nil {
		t.Fatalf("FindAll() got = nil, want not nil")
	}
}
