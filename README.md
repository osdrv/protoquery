# ProtoQuery

## About

ProtoQuery is a Golang library that allows you to traverse a protobuf structure using a simple XPath-like query language.

## Status

This library is in early development and is not yet ready for production use.

## Examples

Given the following protobuf structure:

```protobuf
message AddressBook {
  repeated Contact contacts = 1;
}

message Contact {
  string name = 1;
  int32 id = 2;
  string email = 3;

  enum PhoneType {
    MOBILE = 0;
    HOME = 1;
    WORK = 2;
  }

  message PhoneNumber {
    string number = 1;
    PhoneType type = 2;
  }

  repeated PhoneNumber phones = 4;
}
```

You can query it like this:

```go
package main

import (
	"log"

	"github.com/osdrv/protoquery"
)

func main() {
	q, err := protoquery.Compile("/contacts[@name='John']/phones[@type='WORK']/number")
	if err != nil {
		log.Fatalf("Failed to compile query: %v", err)
	}

	ab, err := loadAddressBook()
	if err != nil {
		log.Fatalf("Failed to load address book: %v", err)
	}

	phones := q.FindAll(ab)
	for _, phone := range phones {
		log.Printf("Phone number: %s", phone)
	}
}

```

## Contributing

Contributions are welcome! Please open a ticket or a pull request.

## License

ProtoQuery is licensed under the MIT license. Please see the LICENSE file for more information.

## Author and Copyright

© Oleg Sidorov, 2024
