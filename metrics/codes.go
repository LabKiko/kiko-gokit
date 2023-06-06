package metrics

import (
	"encoding/json"
	"fmt"
	"strconv"
)

const (
	// Unset is the default status code.
	Unset Code = 0
	// Error indicates the operation contains an error.
	Error Code = 1
	// OK indicates operation has been validated by an Application developers
	// or Operator to have completed successfully, or contain no error.
	OK Code = 2

	// NotFound means some requested entity (e.g., file or directory) was
	// not found.
	//
	// This error code will not be generated by the gRPC framework.
	NotFound Code = 5

	maxCode = 20
)

// Code is an 32-bit representation of a status state.
type Code uint32

var codeToStr = map[Code]string{
	Unset:    "unset",
	Error:    "error",
	OK:       "ok",
	NotFound: "not_found",
}

var strToCode = map[string]Code{
	`"unset"`:   Unset,
	`"error"`:   Error,
	`"ok"`:      OK,
	"not_found": NotFound,
}

// String returns the Code as a string.
func (c Code) String() string {
	return codeToStr[c]
}

// UnmarshalJSON unmarshals b into the Code.
//
// This is based on the functionality in the gRPC codes package:
// https://github.com/grpc/grpc-go/blob/bb64fee312b46ebee26be43364a7a966033521b1/codes/codes.go#L218-L244
func (c *Code) UnmarshalJSON(b []byte) error {
	// From json.Unmarshaler: By convention, to approximate the behavior of
	// Unmarshal itself, Unmarshalers implement UnmarshalJSON([]byte("null")) as
	// a no-op.
	if string(b) == "null" {
		return nil
	}
	if c == nil {
		return fmt.Errorf("nil receiver passed to UnmarshalJSON")
	}

	var x interface{}
	if err := json.Unmarshal(b, &x); err != nil {
		return err
	}
	switch x.(type) {
	case string:
		if jc, ok := strToCode[string(b)]; ok {
			*c = jc
			return nil
		}
		return fmt.Errorf("invalid code: %q", string(b))
	case float64:
		if ci, err := strconv.ParseUint(string(b), 10, 32); err == nil {
			if ci >= maxCode {
				return fmt.Errorf("invalid code: %q", ci)
			}

			*c = Code(ci)
			return nil
		}
		return fmt.Errorf("invalid code: %q", string(b))
	default:
		return fmt.Errorf("invalid code: %q", string(b))
	}
}

// MarshalJSON returns c as the JSON encoding of c.
func (c *Code) MarshalJSON() ([]byte, error) {
	if c == nil {
		return []byte("null"), nil
	}
	str, ok := codeToStr[*c]
	if !ok {
		return nil, fmt.Errorf("invalid code: %d", *c)
	}
	return []byte(fmt.Sprintf("%q", str)), nil
}